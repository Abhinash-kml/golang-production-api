package service

import (
	"errors"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
)

type PostsService interface {
	GetPosts() ([]model.Post, error)
	InsertPosts([]model.Post) error
	UpdatePosts([]model.Post) error
	DeletePosts([]model.Post) error
}

type LocalPostsService struct {
	repo repository.PostsRepository
}

func NewLocalPostsService(repository repository.PostsRepository) *LocalPostsService {
	return &LocalPostsService{
		repo: repository,
	}
}

func (s *LocalPostsService) GetPosts() ([]model.Post, error) {
	posts, err := s.repo.GetPosts()
	if err != nil {
		if errors.Is(err, repository.ErrNoPosts) {
			return nil, ErrOpFailed
		}
	}

	return posts, nil
}

func (s *LocalPostsService) InsertPosts(posts []model.Post) error {
	err := s.repo.InsertPosts(posts)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}

func (s *LocalPostsService) UpdatePosts(posts []model.Post) error {
	err := s.repo.UpdatePosts(posts)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}

func (s *LocalPostsService) DeletePosts(posts []model.Post) error {
	err := s.repo.DeletePosts(posts)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}
