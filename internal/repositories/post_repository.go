package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	oteltracer "go.opentelemetry.io/otel/trace"
)

var ErrNoPosts = errors.New("No posts in repository")

type PostsRepository interface {
	Setup() error

	GetPosts(context.Context) ([]model.Post, error)
	GetById(context.Context, int) (*model.Post, error)
	GetPostsOfUser(context.Context, int) ([]model.Post, error)
	InsertPost(context.Context, model.Post) error
	DeletePost(context.Context, int) error
	UpdatePost(context.Context, int, model.Post) error
	Count() int
}

type InMemoryPostsRepository struct {
	posts  []model.Post
	tracer oteltracer.Tracer
}

func NewInMemoryPostsRepository(tracer oteltracer.Tracer) *InMemoryPostsRepository {
	return &InMemoryPostsRepository{tracer: tracer}
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

func (e *InMemoryPostsRepository) GetPosts(ctx context.Context) ([]model.Post, error) {
	ctx, span := e.tracer.Start(ctx, "GetPosts.Repository")
	defer span.End()

	if len(e.posts) <= 0 {
		return nil, ErrNoPosts
	}

	return e.posts, nil
}

func (e *InMemoryPostsRepository) GetById(ctx context.Context, id int) (*model.Post, error) {
	ctx, span := e.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	for _, value := range e.posts {
		if value.Id == id {
			return &value, nil
		}
	}

	return nil, ErrNoRecord
}

func (e *InMemoryPostsRepository) GetPostsOfUser(ctx context.Context, id int) ([]model.Post, error) {
	ctx, span := e.tracer.Start(ctx, "GetPostsOfUser.Repository")
	defer span.End()

	var posts []model.Post
	for _, value := range e.posts {
		if value.AuthorID == id {
			posts = append(posts, value)
		}
	}

	if len(posts) < 1 {
		return nil, ErrNoPosts
	}

	return posts, nil
}

func (e *InMemoryPostsRepository) InsertPost(ctx context.Context, post model.Post) error {
	ctx, span := e.tracer.Start(ctx, "InsertPost.Repository")
	defer span.End()

	e.posts = append(e.posts, post)

	return nil
}

func (e *InMemoryPostsRepository) DeletePost(ctx context.Context, id int) error {
	ctx, span := e.tracer.Start(ctx, "DeletePost.Repository")
	defer span.End()

	e.posts = slices.DeleteFunc(e.posts, func(post model.Post) bool {
		if post.Id == id {
			return true
		}

		return false
	})

	return nil
}

func (e *InMemoryPostsRepository) UpdatePost(ctx context.Context, id int, post model.Post) error {
	ctx, span := e.tracer.Start(ctx, "UpdatePost.Repository")
	defer span.End()

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
