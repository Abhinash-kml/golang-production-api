package service

import (
	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
)

type CommentService interface {
	GetComments() ([]model.CommentResponseDTO, error)
	InsertComment(model.CommentCreateDTO) error
	DeleteComment(int) error
	UpdateComment(int, model.CommentUpdateDTO) error
}

type LocalCommentService struct {
	repo repository.CommentRepository
}

func NewLocalCommentService(repository repository.CommentRepository) *LocalCommentService {
	return &LocalCommentService{
		repo: repository,
	}
}

func (s *LocalCommentService) GetComments() ([]model.CommentResponseDTO, error) {
	comments, err := s.repo.GetComments()
	if err != nil {
		return nil, ErrOpFailed
	}

	dtos := make([]model.CommentResponseDTO, len(comments))

	for index, value := range comments {
		dtos[index] = ConvertCommentToCommentResponseDTO(&value)
	}

	return dtos, nil
}

func (s *LocalCommentService) InsertComment(comment model.CommentCreateDTO) error {
	newcomment := model.Comment{
		Id:          s.repo.Count() + 1,
		CommenterId: comment.Authorid,
		PostId:      comment.Postid,
		Body:        comment.Body,
	}
	err := s.repo.InsertComment(newcomment)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalCommentService) DeleteComment(id int) error {
	err := s.repo.DeleteComment(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalCommentService) UpdateComment(id int, comment model.CommentUpdateDTO) error {
	// TODO: Fetch and replace old values of unmodifed attributes
	updatedcomment := model.Comment{
		Id:   comment.Id,
		Body: comment.Body,
	}
	err := s.repo.UpdateComment(id, updatedcomment)
	if err != nil {
		return err
	}

	return nil
}

func ConvertCommentToCommentResponseDTO(comment *model.Comment) model.CommentResponseDTO {
	return model.CommentResponseDTO{
		Id:          comment.Id,
		PostID:      comment.PostId,
		CommenterId: comment.CommenterId,
		Body:        comment.Body,
		Likes:       comment.Likes,
	}
}
