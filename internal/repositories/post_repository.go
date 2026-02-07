package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

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
	var posts []model.Post

	file, err := os.OpenFile("./mocks/posts.json", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal("Error opening posts mock file. Error:", err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&posts)

	fmt.Println("Successfully read", len(posts), "posts from mocks")

	e.posts = posts

	return nil
}

func (e *InMemoryPostsRepository) GetPosts() ([]model.Post, error) {
	if len(e.posts) <= 0 {
		return nil, ErrNoPosts
	}

	return e.posts, nil
}

func (e *InMemoryPostsRepository) InsertPosts(posts []model.Post) error {
	if len(posts) <= 0 {
		return ErrZeroLengthSlice
	}

	e.posts = append(e.posts, posts...)

	return nil
}

func (e *InMemoryPostsRepository) DeletePosts(posts []model.Post) error {
	if len(posts) <= 0 {
		return ErrZeroLengthSlice
	}

	e.posts = slices.DeleteFunc(e.posts, func(post model.Post) bool {
		for _, value := range posts {
			if value.Id == post.Id {
				return true
			}
		}

		return false
	})

	return nil
}

func (e *InMemoryPostsRepository) UpdatePosts(posts []model.Post) error {
	if len(posts) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, internal := range e.posts {
		for _, incoming := range posts {
			if internal.Id == incoming.Id {
				internal.Body = incoming.Body
			}
		}
	}

	return nil
}
