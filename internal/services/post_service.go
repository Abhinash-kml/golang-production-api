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

type PostsService interface {
	GetPosts() ([]model.PostResponseDTO, error)
	GetById(int) (*model.PostResponseDTO, error)
	GetPostsOfUser(int) ([]model.PostResponseDTO, error)
	InsertPost(model.PostCreateDTO) error
	UpdatePost(int, model.PostUpdateDTO) error
	DeletePost(int) error
}

type LocalPostsService struct {
	repo  repository.PostsRepository
	cache *redis.Client
}

func NewLocalPostsService(repository repository.PostsRepository, cache *redis.Client) *LocalPostsService {
	return &LocalPostsService{
		repo:  repository,
		cache: cache,
	}
}

func (s *LocalPostsService) GetPosts() ([]model.PostResponseDTO, error) {
	posts, err := s.repo.GetPosts()
	if err != nil {
		if errors.Is(err, repository.ErrNoPosts) {
			return nil, ErrOpFailed
		}
	}

	dtos := make([]model.PostResponseDTO, len(posts))

	for index, value := range posts {
		dtos[index] = ConvertPostToPostResponseDTO(&value)
	}

	return dtos, nil
}

func (s *LocalPostsService) GetById(id int) (*model.PostResponseDTO, error) {
	post, err := s.getPostFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		zap.L().Debug("Cache miss", zap.Int("id", id))

		post, err = s.repo.GetById(id) // Get from db in case of cache miss
		if err != nil {
			return nil, err
		}
		go s.addToCache(post) // Add to cache on a separate goroutine asynchronously
	}

	postReponse := ConvertPostToPostResponseDTO(post)
	return &postReponse, nil
}

func (s *LocalPostsService) GetPostsOfUser(id int) ([]model.PostResponseDTO, error) {
	posts, err := s.repo.GetPostsOfUser(id)
	if err != nil {
		return nil, err
	}

	dtos := make([]model.PostResponseDTO, len(posts))

	for index := range posts {
		dtos[index] = ConvertPostToPostResponseDTO(&posts[index])
	}

	return dtos, nil
}

func (s *LocalPostsService) InsertPost(post model.PostCreateDTO) error {
	newpost := model.Post{
		Id:        s.repo.Count() + 1,
		Title:     post.Title,
		Body:      post.Body,
		CreatorId: post.Authorid,
		Likes:     0,
	}
	err := s.repo.InsertPost(newpost)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Improvement needed
func (s *LocalPostsService) UpdatePost(id int, post model.PostUpdateDTO) error {
	updated := model.Post{
		Id:    post.Id,
		Title: post.Title,
		Body:  post.Body,
	}
	err := s.repo.UpdatePost(id, updated)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalPostsService) DeletePost(id int) error {
	err := s.repo.DeletePost(id)
	if err != nil {
		return err
	}

	return nil
}

func ConvertPostToPostResponseDTO(post *model.Post) model.PostResponseDTO {
	return model.PostResponseDTO{
		Id:       post.Id,
		AuthorId: post.CreatorId,
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
