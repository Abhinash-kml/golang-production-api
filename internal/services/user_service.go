package service

import (
	"context"
	"errors"
	"fmt"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrOpFailed = errors.New("Operation failed")
)

type UserService interface {
	GetUsers() ([]model.UserResponseDTO, error)
	GetById(int) (*model.UserResponseDTO, error)
	InsertUser(model.UserCreateDTO) error
	UpdateUser(int, model.UserUpdateDTO) error
	DeleteUser(int) error
}

type LocalUserService struct {
	repo  repository.UserRepository
	cache *redis.Client
}

func NewLocalUserService(repository repository.UserRepository, cache *redis.Client) *LocalUserService {
	return &LocalUserService{
		repo:  repository,
		cache: cache,
	}
}

func (s *LocalUserService) GetUsers() ([]model.UserResponseDTO, error) {
	users, err := s.repo.GetUsers()
	if err != nil {
		if errors.Is(err, repository.ErrNoUsers) {
			return nil, ErrOpFailed
		}
	}

	dtos := make([]model.UserResponseDTO, len(users))

	for index, value := range users {
		dtos[index] = ConvertUserToUserReponseDTO(&value)
	}

	return users, nil
}

func (s *LocalUserService) GetById(id int) (*model.UserResponseDTO, error) {
	user, err := s.getUserFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		zap.L().Debug("Cache miss", zap.Int("id", id))

		user, err = s.repo.GetById(id) // Get from db in case of case miss
		if err != nil {
			return nil, err
		}
		go s.addToCache(user) // Add to cache on a separate goroutine asynchronously
	}

	userResponse := ConvertUserToUserReponseDTO(user)
	return &userResponse, nil
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

func ConvertUserToUserReponseDTO(user *model.User) model.UserResponseDTO {
	return model.UserResponseDTO{
		Id:      user.Id,
		Name:    user.Name,
		City:    user.City,
		State:   user.State,
		Country: user.Country,
	}
}

func (s *LocalUserService) addToCache(u *model.User) {
	formatedId := fmt.Sprintf("user:%d", u.Id)
	ctx := context.Background()
	s.cache.HSet(ctx, formatedId, u)
}

func (s *LocalUserService) getUserFromCache(id int) (*model.User, error) {
	formatedId := fmt.Sprintf("user:%d", id)
	ctx := context.Background()
	user := new(model.User)
	err := s.cache.HGetAll(ctx, formatedId).Scan(user)
	if err != nil {
		return nil, err
	}

	// Manual check for cache miss
	// On cache miss - populate cache with data from db on service layer
	if user.Id == 0 {
		return nil, redis.Nil
	}

	zap.L().Debug("Cache hit", zap.Int("id", id))

	return user, nil
}
