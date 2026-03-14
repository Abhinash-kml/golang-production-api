package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/redis/go-redis/v9"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type CommentService interface {
	GetComments(context.Context) ([]model.CommentResponseDTO, error)
	GetById(context.Context, int) (*model.CommentResponseDTO, error)
	GetCommentsOfPost(context.Context, int) ([]model.CommentResponseDTO, error)
	InsertComment(context.Context, model.CommentCreateDTO) error
	DeleteComment(context.Context, int) error
	UpdateComment(context.Context, int, model.CommentUpdateDTO) error
}

type LocalCommentService struct {
	repo   repository.CommentRepository
	cache  *redis.Client
	tracer oteltracer.Tracer
}

func NewLocalCommentService(repository repository.CommentRepository, conn *connections.RedisConnection, tracer oteltracer.Tracer) *LocalCommentService {
	return &LocalCommentService{
		repo:   repository,
		cache:  conn.Client,
		tracer: tracer,
	}
}

func (s *LocalCommentService) GetComments(ctx context.Context) ([]model.CommentResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetComments.Service")
	defer span.End()

	comments, err := s.repo.GetComments(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOpFailed
		}
	}

	dtos := make([]model.CommentResponseDTO, len(comments))

	for index, value := range comments {
		dtos[index] = ConvertCommentToCommentResponseDTO(&value)
	}

	return dtos, nil
}

func (s *LocalCommentService) GetById(ctx context.Context, id int) (*model.CommentResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetById.Service")
	defer span.End()

	comment, err := s.getCommentFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		zap.L().Debug("Cache miss", zap.Int("id", id))

		comment, err = s.repo.GetById(ctx, id)
		if err != nil {
			if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
				return nil, ErrOpFailed
			}
		}
		go s.addToCache(comment) // Add to cache on a separate goroutine asynchronously
	}

	commentResponse := ConvertCommentToCommentResponseDTO(comment)
	return &commentResponse, nil
}

func (s *LocalCommentService) GetCommentsOfPost(ctx context.Context, id int) ([]model.CommentResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetCommentsOfPost.Service")
	defer span.End()

	comments, err := s.repo.GetCommentsOfPost(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOpFailed
		}
	}

	dtos := make([]model.CommentResponseDTO, len(comments))
	for index := range comments {
		dtos[index] = ConvertCommentToCommentResponseDTO(&comments[index])
	}

	return dtos, nil
}

func (s *LocalCommentService) InsertComment(ctx context.Context, comment model.CommentCreateDTO) error {
	ctx, span := s.tracer.Start(ctx, "InsertComment.Service")
	defer span.End()

	newcomment := model.Comment{
		Id:       s.repo.Count() + 1,
		AuthorID: comment.Authorid,
		PostId:   comment.Postid,
		Body:     comment.Body,
	}
	err := s.repo.InsertComment(ctx, newcomment)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalCommentService) DeleteComment(ctx context.Context, id int) error {
	ctx, span := s.tracer.Start(ctx, "DeleteComment.Service")
	defer span.End()

	err := s.repo.DeleteComment(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalCommentService) UpdateComment(ctx context.Context, id int, comment model.CommentUpdateDTO) error {
	ctx, span := s.tracer.Start(ctx, "UpdateComment.Service")
	defer span.End()

	// TODO: Fetch and replace old values of unmodifed attributes
	updatedcomment := model.Comment{
		Id:   comment.Id,
		Body: comment.Body,
	}
	err := s.repo.UpdateComment(ctx, id, updatedcomment)
	if err != nil {
		return err
	}

	return nil
}

func ConvertCommentToCommentResponseDTO(comment *model.Comment) model.CommentResponseDTO {
	return model.CommentResponseDTO{
		Id:       comment.Id,
		PostID:   comment.PostId,
		AuthorID: comment.AuthorID,
		Body:     comment.Body,
		Likes:    comment.Likes,
	}
}

func (s *LocalCommentService) addToCache(u *model.Comment) {
	formatedId := fmt.Sprintf("comment:%d", u.Id)
	ctx := context.Background()
	s.cache.HSet(ctx, formatedId, u)
}

func (s *LocalCommentService) getCommentFromCache(id int) (*model.Comment, error) {
	formatedId := fmt.Sprintf("comment:%d", id)
	ctx := context.Background()
	comment := new(model.Comment)
	err := s.cache.HGetAll(ctx, formatedId).Scan(comment)
	if err != nil {
		return nil, err
	}

	// Manual check for cache miss
	// On cache miss - populate cache with data from db o nservice layer
	if comment.Id == 0 {
		return nil, redis.Nil
	}

	zap.L().Debug("Cache hit", zap.Int("id", id))

	return comment, nil
}
