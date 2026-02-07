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

var (
	ErrNoComments         = errors.New("No user in repository")
	ErrUndefinedComments  = errors.New("Undefined users")
	ErrCommentDoesntExist = errors.New("Provided users doesn't exist")
)

type CommentRepository interface {
	Setup() error

	GetComments() ([]model.Comment, error)
	InsertComments([]model.Comment) error
	DeleteComments([]model.Comment) error
	UpdateComments([]model.Comment) error
}

type InMemoryCommentRepository struct {
	comments []model.Comment
}

func NewInMemoryCommentsRepository() *InMemoryCommentRepository {
	return &InMemoryCommentRepository{}
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

func (e *InMemoryCommentRepository) GetComments() ([]model.Comment, error) {
	if len(e.comments) <= 0 {
		return nil, ErrNoComments
	}

	return e.comments, nil
}

func (e *InMemoryCommentRepository) InsertComments(comments []model.Comment) error {
	if len(comments) <= 0 {
		return ErrZeroLengthSlice
	}

	e.comments = append(e.comments, comments...)
	return nil
}

func (e *InMemoryCommentRepository) DeleteComments(comments []model.Comment) error {
	if len(comments) <= 0 {
		return ErrZeroLengthSlice
	}

	if len(e.comments) <= 0 {
		return ErrNoComments
	}

	e.comments = slices.DeleteFunc(e.comments, func(e model.Comment) bool {
		for _, value := range comments {
			if value.Id == e.Id {
				return true
			}
		}

		return false
	})

	return nil
}

func (e *InMemoryCommentRepository) UpdateComments(comments []model.Comment) error {
	if len(comments) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, internal := range e.comments {
		for _, incoming := range comments {
			if internal.Body == incoming.Body {
				internal.Body = incoming.Body
			}
		}
	}

	return nil
}
