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
	InsertPost(model.Post) error
	DeletePost(int) error
	UpdatePost(int, model.Post) error
	Count() int
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

func (e *InMemoryPostsRepository) InsertPost(post model.Post) error {
	e.posts = append(e.posts, post)

	return nil
}

func (e *InMemoryPostsRepository) DeletePost(id int) error {
	e.posts = slices.DeleteFunc(e.posts, func(post model.Post) bool {
		if post.Id == id {
			return true
		}

		return false
	})

	return nil
}

func (e *InMemoryPostsRepository) UpdatePost(id int, post model.Post) error {
	for index := range e.posts {
		if e.posts[index].Id == id {
			e.posts[index] = post
			break
		}
	}

	return nil
}

func (e *InMemoryPostsRepository) Count() int {
	return len(e.posts)
}
