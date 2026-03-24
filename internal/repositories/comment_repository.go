package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type CommentRepository interface {
	Setup() error

	GetComments(context.Context) ([]model.Comment, error)
	GetById(context.Context, int) (*model.Comment, error)
	GetCommentsOfPost(context.Context, int) ([]model.Comment, error)
	InsertComment(context.Context, model.Comment) error
	DeleteComment(context.Context, int) error
	UpdateComment(context.Context, model.CommentUpdateDTO) error
	ReplaceComment(context.Context, model.CommentReplaceDTO) error
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
		span.SetAttributes(attribute.Bool("comments.found", false))
		span.SetStatus(codes.Error, "failed to fetch comments in repository")
		return nil, ErrNoRecord
	}

	span.SetAttributes(attribute.Bool("comments.found", true), attribute.Int("comments.num", len(e.comments)))

	return e.comments, nil
}

func (e *InMemoryCommentRepository) GetById(ctx context.Context, id int) (*model.Comment, error) {
	ctx, span := e.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("comment.id", id))

	for _, value := range e.comments {
		if value.Id == id {
			return &value, nil
		}
	}

	span.SetAttributes(attribute.Bool("comment.found", false))
	return nil, ErrNoRecord
}

func (e *InMemoryCommentRepository) GetCommentsOfPost(ctx context.Context, id int) ([]model.Comment, error) {
	ctx, span := e.tracer.Start(ctx, "GetCommentsOfPost.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("post.id", id))

	var comments []model.Comment

	for _, value := range e.comments {
		if value.PostId == id {
			comments = append(comments, value)
		}
	}

	if len(comments) < 1 {
		span.SetAttributes(attribute.Bool("comments.found", false))
		span.SetStatus(codes.Error, "failed to fetch comments of post in repository")
		return nil, ErrNoRecord
	}

	span.SetAttributes(attribute.Bool("comments.found", true))

	return comments, nil
}

func (e *InMemoryCommentRepository) InsertComment(ctx context.Context, comment model.Comment) error {
	ctx, span := e.tracer.Start(ctx, "InsertComment.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("comment.id", comment.Id),
		attribute.Int("comment.postid", comment.PostId),
		attribute.Int("comment.authorid", comment.AuthorID),
		attribute.String("comment.body", comment.Body))

	e.comments = append(e.comments, comment)

	span.SetAttributes(attribute.Bool("comment.inserted", true))
	return nil
}

func (e *InMemoryCommentRepository) DeleteComment(ctx context.Context, id int) error {
	ctx, span := e.tracer.Start(ctx, "DeleteComment.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("comment.id", id))

	oldlen := len(e.comments)
	e.comments = slices.DeleteFunc(e.comments, func(e model.Comment) bool {
		if e.Id == id {
			return true
		}

		return false
	})
	newlen := len(e.comments)

	if newlen != oldlen {
		span.SetAttributes(attribute.Bool("comment.deleted", true))
	} else {
		span.SetAttributes(attribute.Bool("comment.deleted", false))
	}

	return nil
}

// TODO: Implement as per JSON Merge patch
func (e *InMemoryCommentRepository) UpdateComment(ctx context.Context, dto model.CommentUpdateDTO) error {
	ctx, span := e.tracer.Start(ctx, "UpdateComment.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("comment.id", dto.Id),
		attribute.Int("comment.patch.num", len(dto.Patches)))

	span.SetAttributes(attribute.Int("comment.id", dto.Id),
		attribute.Int("comment.patch.num", len(dto.Patches)))

	var comment *model.Comment
	for index := range e.comments {
		if e.comments[index].Id == dto.Id {
			comment = &e.comments[index]
			break
		}
	}

	for index := range dto.Patches {
		current := dto.Patches[index]
		field := current.Field
		interfaceValue := current.Value

		switch field {
		case "body":
			value, ok := interfaceValue.(string)
			if !ok {
				zap.L().Fatal("Failed to update comment due to type assertion", zap.String("field", "body"))
			}
			comment.Body = value
		}
	}

	// Update comment in cache

	return nil
}

func (e *InMemoryCommentRepository) ReplaceComment(ctx context.Context, dto model.CommentReplaceDTO) error {
	return nil
}

func (e *InMemoryCommentRepository) Count() int {
	return len(e.comments)
}
