package repository

import (
	"errors"

	model "github.com/abhinash-kml/go-api-server/internal/models"
)

var ErrNoPosts = errors.New("No posts in repository")

type PostsRepository interface {
	Setup() error

	GetPosts() ([]model.Post, error)
	InsertPosts([]model.Post) error
	DeletePosts([]model.Post) error
	UpdatePosts([]model.Post) error
}

type InMemoryPostsRepository struct {
	posts []model.Post
}

func NewInMemoryPostsRepository() *InMemoryPostsRepository {
	return &InMemoryPostsRepository{}
}

func (e *InMemoryPostsRepository) Setup() error {
	return nil
}

func (e *InMemoryPostsRepository) GetPosts() ([]model.Post, error) {
	return nil, nil
}

func (e *InMemoryPostsRepository) InsertPosts(posts []model.Post) error {
	return nil
}

func (e *InMemoryPostsRepository) DeletePosts(posts []model.Post) error {
	return nil
}

func (e *InMemoryPostsRepository) UpdatePosts(posts []model.Post) error {
	return nil
}
