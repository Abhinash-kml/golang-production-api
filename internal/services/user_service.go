package service

import (
	"errors"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
)

var (
	ErrOpFailed = errors.New("Operation failed")
)

type UserService interface {
	GetUsers() ([]model.User, error)
	InsertUsers([]model.User) error
	UpdateUsers([]model.User, []model.User) error
	DeleteUsers([]model.User) error
}

type LocalUserService struct {
	repo repository.UserRepository
}

func NewLocalUserService(repository repository.UserRepository) *LocalUserService {
	return &LocalUserService{
		repo: repository,
	}
}

func (s *LocalUserService) GetUsers() ([]model.User, error) {
	users, err := s.repo.GetUsers()
	if err != nil {
		if errors.Is(err, repository.ErrNoUsers) {
			return nil, ErrOpFailed
		}
	}

	return users, nil
}

func (s *LocalUserService) InsertUsers(users []model.User) error {
	err := s.repo.InsertUsers(users)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}

func (s *LocalUserService) UpdateUsers(old, new []model.User) error {
	err := s.repo.UpdateUsers(old, new)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}

func (s *LocalUserService) DeleteUsers(users []model.User) error {
	err := s.repo.DeleteUsers(users)
	if err != nil {
		if errors.Is(err, repository.ErrZeroLengthSlice) {
			return ErrOpFailed
		}
	}

	return nil
}
