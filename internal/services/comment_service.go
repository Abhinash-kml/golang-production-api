package service

import (
	"context"
	"errors"
	"fmt"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CommentService interface {
	GetComments() ([]model.CommentResponseDTO, error)
	GetById(int) (*model.CommentResponseDTO, error)
	GetCommentsOfPost(int) ([]model.CommentResponseDTO, error)
	InsertComment(model.CommentCreateDTO) error
	DeleteComment(int) error
	UpdateComment(int, model.CommentUpdateDTO) error
}

type LocalCommentService struct {
	repo  repository.CommentRepository
	cache *redis.Client
}

func NewLocalCommentService(repository repository.CommentRepository, cache *redis.Client) *LocalCommentService {
	return &LocalCommentService{
		repo:  repository,
		cache: cache,
	}
}

func (s *LocalCommentService) GetComments() ([]model.CommentResponseDTO, error) {
	comments, err := s.repo.GetComments()
	if err != nil {
		return nil, ErrOpFailed
	}

	dtos := make([]model.CommentResponseDTO, len(comments))

	for index, value := range comments {
		dtos[index] = ConvertCommentToCommentResponseDTO(&value)
	}

	return dtos, nil
}

func (s *LocalCommentService) GetById(id int) (*model.CommentResponseDTO, error) {
	comment, err := s.getCommentFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		zap.L().Debug("Cache miss", zap.Int("id", id))

		comment, err = s.repo.GetById(id)
		if err != nil {
			return nil, err
		}
	}

	commentResponse := ConvertCommentToCommentResponseDTO(comment)
	return &commentResponse, nil
}

func (s *LocalCommentService) GetCommentsOfPost(id int) ([]model.CommentResponseDTO, error) {
	comments, err := s.repo.GetCommentsOfPost(id)
	if err != nil {
		return nil, err
	}

	dtos := make([]model.CommentResponseDTO, len(comments))
	for index := range comments {
		dtos[index] = ConvertCommentToCommentResponseDTO(&comments[index])
	}

	return dtos, nil
}

func (s *LocalCommentService) InsertComment(comment model.CommentCreateDTO) error {
	newcomment := model.Comment{
		Id:          s.repo.Count() + 1,
		CommenterId: comment.Authorid,
		PostId:      comment.Postid,
		Body:        comment.Body,
	}
	err := s.repo.InsertComment(newcomment)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalCommentService) DeleteComment(id int) error {
	err := s.repo.DeleteComment(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalCommentService) UpdateComment(id int, comment model.CommentUpdateDTO) error {
	// TODO: Fetch and replace old values of unmodifed attributes
	updatedcomment := model.Comment{
		Id:   comment.Id,
		Body: comment.Body,
	}
	err := s.repo.UpdateComment(id, updatedcomment)
	if err != nil {
		return err
	}

	return nil
}

func ConvertCommentToCommentResponseDTO(comment *model.Comment) model.CommentResponseDTO {
	return model.CommentResponseDTO{
		Id:          comment.Id,
		PostID:      comment.PostId,
		CommenterId: comment.CommenterId,
		Body:        comment.Body,
		Likes:       comment.Likes,
	}
}

func (s *LocalCommentService) getCommentFromCache(id int) (*model.Comment, error) {
	formatedId := fmt.Sprintf("comment:%d", id)
	ctx := context.Background()
	comment := new(model.Comment)
	err := s.cache.HGetAll(ctx, formatedId).Scan(comment)
	if err != nil {
		return nil, err
	}

	// lambda to add to cache
	AddToCache := func() {
		commentFromDB, err := s.repo.GetById(id)
		if err != nil {
			return
		}

		s.cache.HSet(ctx, formatedId, commentFromDB)
	}

	// Manual check for cache miss
	// On cache miss - populate cache with data from db
	if comment.Id == 0 {
		go AddToCache() // Add to cache
		return nil, redis.Nil
	}

	zap.L().Debug("Cache hit", zap.Int("id", id))

	return comment, nil
}
