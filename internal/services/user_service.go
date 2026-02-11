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
	GetUsers() ([]model.UserResponseDTO, error)
	InsertUser(model.UserCreateDTO) error
	UpdateUser(int, model.UserUpdateDTO) error
	DeleteUser(int) error
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

func (s *LocalUserService) InsertUser(user model.UserCreateDTO) error {
	newuser := model.User{
		Id:      s.repo.Count() + 1,
		Name:    user.Name,
		City:    user.City,
		State:   user.State,
		Country: user.Country,
	}
	err := s.repo.InsertUser(newuser)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalUserService) UpdateUser(id int, new model.UserUpdateDTO) error {
	updateduser := model.User{
		Id: id,
	}

	// TODO: Improve this
	switch new.What {
	case "name":
		{
			updateduser.Name = new.NewData
		}
	case "country":
		{
			updateduser.Country = new.NewData
		}
	case "city":
		{
			updateduser.City = new.NewData
		}
	case "state":
		{
			updateduser.State = new.NewData
		}
	}

	err := s.repo.UpdateUser(id, updateduser)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalUserService) DeleteUser(id int) error {
	err := s.repo.DeleteUser(id)
	if err != nil {
		return err
	}

	return nil
}
