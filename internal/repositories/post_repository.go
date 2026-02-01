package repository

import model "github.com/abhinash-kml/go-api-server/internal/models"

type PostRepository interface {
	Setup() error

	GetPosts() ([]model.Post, error)
	InsertPosts([]model.Post) error
	DeletePosts([]model.Post) error
	UpdatePosts([]model.Post) error
}

type InMemoryPostRepository struct {
	posts []model.Post
}

func (e *InMemoryPostRepository) Setup() error {
	return nil
}

func (e *InMemoryPostRepository) GetPosts() ([]model.Post, error) {
	return nil, nil
}

func (e *InMemoryPostRepository) InsertPosts(posts []model.Post) error {
	return nil
}

func (e *InMemoryPostRepository) DeletePosts(posts []model.Post) error {
	return nil
}

func (e *InMemoryPostRepository) UpdatePosts(posts []model.Post) error {
	return nil
}
