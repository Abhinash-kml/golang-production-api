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
	GetById(int) (*model.Comment, error)
	GetCommentsOfPost(int) ([]model.Comment, error)
	InsertComment(model.Comment) error
	DeleteComment(int) error
	UpdateComment(int, model.Comment) error
	Count() int
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

func (e *InMemoryCommentRepository) GetById(id int) (*model.Comment, error) {
	for _, value := range e.comments {
		if value.Id == id {
			return &value, nil
		}
	}

	return nil, ErrNoRecord
}

func (e *InMemoryCommentRepository) GetCommentsOfPost(id int) ([]model.Comment, error) {
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

func (e *InMemoryCommentRepository) InsertComment(comment model.Comment) error {
	e.comments = append(e.comments, comment)
	return nil
}

func (e *InMemoryCommentRepository) DeleteComment(id int) error {
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

func (e *InMemoryCommentRepository) UpdateComment(id int, comment model.Comment) error {
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
