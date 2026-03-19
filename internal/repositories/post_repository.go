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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	UpdatePost(context.Context, model.PostUpdateDTO) error
	ReplacePost(context.Context, model.PostReplaceDTO) error
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
		span.SetAttributes(attribute.Bool("posts.found", false))
		span.SetStatus(codes.Error, "failed to fetch posts in repository")
		return nil, ErrNoRecord
	}

	span.SetAttributes(attribute.Bool("posts.found", true), attribute.Int("posts.num", len(e.posts)))

	return e.posts, nil
}

func (e *InMemoryPostsRepository) GetById(ctx context.Context, id int) (*model.Post, error) {
	ctx, span := e.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", id))

	for _, value := range e.posts {
		if value.Id == id {
			return &value, nil
		}
	}

	span.SetAttributes(attribute.Bool("post.found", false))
	span.SetStatus(codes.Error, "failed to fetch post in repository")
	return nil, ErrNoRecord
}

func (e *InMemoryPostsRepository) GetPostsOfUser(ctx context.Context, id int) ([]model.Post, error) {
	ctx, span := e.tracer.Start(ctx, "GetPostsOfUser.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", id))

	var posts []model.Post
	for _, value := range e.posts {
		if value.AuthorID == id {
			posts = append(posts, value)
		}
	}

	if len(posts) < 1 {
		span.SetAttributes(attribute.Bool("posts.found", false))
		span.SetStatus(codes.Error, "failed to fetch posts of user in repository")
		return nil, ErrNoRecord
	}

	span.SetAttributes(attribute.Bool("posts.found", true), attribute.Int("posts.num", len(posts)))
	return posts, nil
}

func (e *InMemoryPostsRepository) InsertPost(ctx context.Context, post model.Post) error {
	ctx, span := e.tracer.Start(ctx, "InsertPost.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", post.Id),
		attribute.Int("post.authorid", post.AuthorID),
		attribute.String("post.title", post.Title),
		attribute.String("post.body", post.Body),
		attribute.Int("post.likes", post.Likes),
		attribute.Int("post.created_at", int(post.CreatedAt.Unix())))

	e.posts = append(e.posts, post)

	span.SetAttributes(attribute.Bool("post.inserted", true))

	return nil
}

func (e *InMemoryPostsRepository) DeletePost(ctx context.Context, id int) error {
	ctx, span := e.tracer.Start(ctx, "DeletePost.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", id))

	oldlen := len(e.posts)
	e.posts = slices.DeleteFunc(e.posts, func(post model.Post) bool {
		if post.Id == id {
			return true
		}

		return false
	})
	newlen := len(e.posts)

	if newlen != oldlen {
		span.SetAttributes(attribute.Bool("post.deleted", true))
	} else {
		span.SetAttributes(attribute.Bool("post.deleted", false))
	}

	return nil
}

// TODO: Implement after thinking
func (e *InMemoryPostsRepository) UpdatePost(ctx context.Context, dto model.PostUpdateDTO) error {
	ctx, span := e.tracer.Start(ctx, "UpdatePost.Repository")
	defer span.End()

	for index := range e.posts {
		if e.posts[index].Id == dto.Id {
			// e.posts[index] = model.Post(dto)
			break
		}
	}

	return nil
}

func (e *InMemoryPostsRepository) ReplacePost(ctx context.Context, dto model.PostReplaceDTO) error {
	ctx, span := e.tracer.Start(ctx, "UpdatePost.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", dto.Id),
		attribute.String("post.title", dto.Title),
		attribute.String("post.body", dto.Body))

	return nil
}

func (e *InMemoryPostsRepository) Count() int {
	return len(e.posts)
}
