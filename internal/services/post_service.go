package service

import (
	"errors"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
)

type PostsService interface {
	GetPosts() ([]model.PostResponseDTO, error)
	InsertPost(model.PostCreateDTO) error
	UpdatePost(int, model.PostUpdateDTO) error
	DeletePost(int) error
}

type LocalPostsService struct {
	repo repository.PostsRepository
}

func NewLocalPostsService(repository repository.PostsRepository) *LocalPostsService {
	return &LocalPostsService{
		repo: repository,
	}
}

func (s *LocalPostsService) GetPosts() ([]model.PostResponseDTO, error) {
	posts, err := s.repo.GetPosts()
	if err != nil {
		if errors.Is(err, repository.ErrNoPosts) {
			return nil, ErrOpFailed
		}
	}

	dtos := make([]model.PostResponseDTO, len(posts))

	for index, value := range posts {
		dtos[index] = ConvertPostToPostResponseDTO(&value)
	}

	return dtos, nil
}

func (s *LocalPostsService) InsertPost(post model.PostCreateDTO) error {
	newpost := model.Post{
		Id:        s.repo.Count() + 1,
		Title:     post.Title,
		Body:      post.Body,
		CreatorId: post.Authorid,
		Likes:     0,
	}
	err := s.repo.InsertPost(newpost)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Improvement needed
func (s *LocalPostsService) UpdatePost(id int, post model.PostUpdateDTO) error {
	updated := model.Post{
		Id:    post.Id,
		Title: post.Title,
		Body:  post.Body,
	}
	err := s.repo.UpdatePost(id, updated)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalPostsService) DeletePost(id int) error {
	err := s.repo.DeletePost(id)
	if err != nil {
		return err
	}

	return nil
}

func ConvertPostToPostResponseDTO(post *model.Post) model.PostResponseDTO {
	return model.PostResponseDTO{
		Id:       post.Id,
		AuthorId: post.CreatorId,
		Title:    post.Title,
		Body:     post.Body,
		Likes:    post.Likes,
		Comments: 0, // TODO: Later
	}
}
