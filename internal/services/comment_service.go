package service

import (
	"errors"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
)

type CommentService interface {
	GetComments() ([]model.Comment, error)
	InsertComments([]model.Comment) error
	DeleteComments([]model.Comment) error
	UpdateComments([]model.Comment) error
}

type LocalCommentService struct {
	repo repository.CommentRepository
}

func NewLocalCommentService(repository repository.CommentRepository) *LocalCommentService {
	return &LocalCommentService{
		repo: repository,
	}
}

func (s *LocalCommentService) GetComments() ([]model.Comment, error) {
	comments, err := s.repo.GetComments()
	if err != nil {
		return nil, ErrOpFailed
	}

	return comments, nil
}

func (s *LocalCommentService) InsertComments(comments []model.Comment) error {
	err := s.repo.InsertComments(comments)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}

func (s *LocalCommentService) DeleteComments(comments []model.Comment) error {
	err := s.repo.DeleteComments(comments)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}

func (s *LocalCommentService) UpdateComments(comments []model.Comment) error {
	err := s.repo.UpdateComments(comments)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}
