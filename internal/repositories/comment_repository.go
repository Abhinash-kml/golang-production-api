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

var (
	ErrNoComments         = errors.New("No user in repository")
	ErrUndefinedComments  = errors.New("Undefined users")
	ErrCommentDoesntExist = errors.New("Provided users doesn't exist")
)

type CommentRepository interface {
	Setup() error

	GetComments(context.Context) ([]model.Comment, error)
	GetById(context.Context, int) (*model.Comment, error)
	GetCommentsOfPost(context.Context, int) ([]model.Comment, error)
	InsertComment(context.Context, model.Comment) error
	DeleteComment(context.Context, int) error
	UpdateComment(context.Context, int, model.Comment) error
	Count() int
}

type InMemoryCommentRepository struct {
	comments []model.Comment
	tracer   oteltracer.Tracer
}

func NewInMemoryCommentsRepository(tracer oteltracer.Tracer) *InMemoryCommentRepository {
	return &InMemoryCommentRepository{tracer: tracer}
}

func (e *InMemoryCommentRepository) Setup() error {
	var comments []model.Comment

	file, err := os.OpenFile("./mocks/comments.json", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open comments mock data file. Error:", err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&comments)

	fmt.Println("Successfully read", len(comments), "comments from mocks")

	// Assign fake to real container ( this is safe because of heap escape )
	e.comments = comments

	return nil
}

func (e *InMemoryCommentRepository) GetComments(ctx context.Context) ([]model.Comment, error) {
	ctx, span := e.tracer.Start(ctx, "GetComments.Repository")
	defer span.End()

	if len(e.comments) <= 0 {
		return nil, ErrNoComments
	}

	return e.comments, nil
}

func (e *InMemoryCommentRepository) GetById(ctx context.Context, id int) (*model.Comment, error) {
	ctx, span := e.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	for _, value := range e.comments {
		if value.Id == id {
			return &value, nil
		}
	}

	return nil, ErrNoRecord
}

func (e *InMemoryCommentRepository) GetCommentsOfPost(ctx context.Context, id int) ([]model.Comment, error) {
	ctx, span := e.tracer.Start(ctx, "GetCommentsOfPost.Repository")
	defer span.End()

	var comments []model.Comment

	for _, value := range e.comments {
		if value.PostId == id {
			comments = append(comments, value)
		}
	}

	if len(comments) < 1 {
		return nil, ErrNoComments
	}

	return comments, nil
}

func (e *InMemoryCommentRepository) InsertComment(ctx context.Context, comment model.Comment) error {
	ctx, span := e.tracer.Start(ctx, "InsertComment.Repository")
	defer span.End()

	e.comments = append(e.comments, comment)
	return nil
}

func (e *InMemoryCommentRepository) DeleteComment(ctx context.Context, id int) error {
	ctx, span := e.tracer.Start(ctx, "DeleteComment.Repository")
	defer span.End()

	if len(e.comments) <= 0 {
		return ErrNoComments
	}

	e.comments = slices.DeleteFunc(e.comments, func(e model.Comment) bool {
		if e.Id == id {
			return true
		}

		return false
	})

	return nil
}

func (e *InMemoryCommentRepository) UpdateComment(ctx context.Context, id int, comment model.Comment) error {
	ctx, span := e.tracer.Start(ctx, "UpdateComment.Repository")
	defer span.End()

	for index := range e.comments {
		if e.comments[index].Id == id {
			e.comments[index] = comment
			break
		}
	}

	return nil
}

func (e *InMemoryCommentRepository) Count() int {
	return len(e.comments)
}
