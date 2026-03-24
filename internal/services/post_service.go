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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type PostsService interface {
	GetPosts(context.Context) ([]model.PostResponseDTO, error)
	GetById(context.Context, int) (*model.PostResponseDTO, error)
	GetPostsOfUser(context.Context, int) ([]model.PostResponseDTO, error)
	InsertPost(context.Context, model.PostCreateDTO) error
	UpdatePost(context.Context, model.PostUpdateDTO) error
	ReplacePost(context.Context, model.PostReplaceDTO) error
	DeletePost(context.Context, int) error
}

type LocalPostsService struct {
	repo   repository.PostsRepository
	cache  *redis.Client
	tracer oteltracer.Tracer
}

func NewLocalPostsService(repository repository.PostsRepository, conn *connections.RedisConnection, tracer oteltracer.Tracer) *LocalPostsService {
	return &LocalPostsService{
		repo:   repository,
		cache:  conn.Client,
		tracer: tracer,
	}
}

func (s *LocalPostsService) GetPosts(ctx context.Context) ([]model.PostResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetPosts.Service")
	defer span.End()

	posts, err := s.repo.GetPosts(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error get posts from repository")
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoRecord
		} else {
			return nil, err
		}
	}

	dtos := make([]model.PostResponseDTO, len(posts))

	for index, value := range posts {
		dtos[index] = ConvertPostToPostResponseDTO(&value)
	}

	return dtos, nil
}

func (s *LocalPostsService) GetById(ctx context.Context, id int) (*model.PostResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetById.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", id))

	post, err := s.getPostFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		span.RecordError(err)
		zap.L().Debug("Cache miss", zap.Int("id", id))

		post, err = s.repo.GetById(ctx, id) // Get from db in case of cache miss
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to fetch post from repository")
			if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNoRecord
			} else {
				return nil, err
			}
		}
		go s.addToCache(post) // Add to cache on a separate goroutine asynchronously
	}

	postReponse := ConvertPostToPostResponseDTO(post)
	return &postReponse, nil
}

func (s *LocalPostsService) GetPostsOfUser(ctx context.Context, id int) ([]model.PostResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetPostsOfUser.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", id))

	posts, err := s.repo.GetPostsOfUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch posts of user from repository")
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoRecord
		} else {
			return nil, err
		}
	}

	dtos := make([]model.PostResponseDTO, len(posts))

	for index := range posts {
		dtos[index] = ConvertPostToPostResponseDTO(&posts[index])
	}

	return dtos, nil
}

func (s *LocalPostsService) InsertPost(ctx context.Context, post model.PostCreateDTO) error {
	ctx, span := s.tracer.Start(ctx, "InsertPost.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("post.authorid", post.AuthorID),
		attribute.String("post.title", post.Title),
		attribute.String("post.body", post.Body))

	newpost := model.Post{
		Id:       s.repo.Count() + 1,
		Title:    post.Title,
		Body:     post.Body,
		AuthorID: post.AuthorID,
		Likes:    0,
	}
	err := s.repo.InsertPost(ctx, newpost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert new post in repository")
		return err
	}

	return nil
}

// TODO: Implement as per JSON merge patch
func (s *LocalPostsService) UpdatePost(ctx context.Context, dto model.PostUpdateDTO) error {
	ctx, span := s.tracer.Start(ctx, "UpdatePost.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", dto.Id),
		attribute.Int("post.patch.num", len(dto.Patches)))

	err := s.repo.UpdatePost(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update post in repository")
		return err
	}

	return nil
}

func (s *LocalPostsService) ReplacePost(ctx context.Context, dto model.PostReplaceDTO) error {
	ctx, span := s.tracer.Start(ctx, "ReplacePost.Service")
	defer span.End()

	err := s.repo.ReplacePost(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to replace post in repository")
		return err
	}

	return nil
}

func (s *LocalPostsService) DeletePost(ctx context.Context, id int) error {
	ctx, span := s.tracer.Start(ctx, "GetUsers.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", id))

	err := s.repo.DeletePost(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete post in repository")
		return err
	}

	return nil
}

func ConvertPostToPostResponseDTO(post *model.Post) model.PostResponseDTO {
	return model.PostResponseDTO{
		Id:       post.Id,
		AuthorId: post.AuthorID,
		Title:    post.Title,
		Body:     post.Body,
		Likes:    post.Likes,
		Comments: 0, // TODO: Later
	}
}

func (s *LocalPostsService) addToCache(u *model.Post) {
	formatedId := fmt.Sprintf("post:%d", u.Id)
	ctx := context.Background()
	s.cache.HSet(ctx, formatedId, u)
}

func (s *LocalPostsService) getPostFromCache(id int) (*model.Post, error) {
	formatedId := fmt.Sprintf("post:%d", id)
	ctx := context.Background()
	post := new(model.Post)
	err := s.cache.HGetAll(ctx, formatedId).Scan(post)
	if err != nil {
		return nil, err
	}

	// On cache miss - populate cache with data from db o nservice layer
	if post.Id == 0 {
		return nil, redis.Nil
	}

	zap.L().Debug("Cache hit", zap.Int("id", id))

	return post, nil
}
