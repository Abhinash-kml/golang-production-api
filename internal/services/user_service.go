package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/redis/go-redis/v9"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	ErrOpFailed = errors.New("Operation failed")
)

type UserService interface {
	GetUsers(context.Context) ([]model.UserResponseDTO, error)
	GetById(context.Context, int) (*model.UserResponseDTO, error)
	InsertUser(context.Context, model.UserCreateDTO) error
	UpdateUser(context.Context, int, model.UserUpdateDTO) error
	DeleteUser(context.Context, int) error
}

type LocalUserService struct {
	repo   repository.UserRepository
	cache  *redis.Client
	tracer oteltracer.Tracer
}

func NewLocalUserService(repository repository.UserRepository, conn *connections.RedisConnection, tracer oteltracer.Tracer) *LocalUserService {
	return &LocalUserService{
		repo:   repository,
		cache:  conn.Client,
		tracer: tracer,
	}
}

func (s *LocalUserService) GetUsers(ctx context.Context) ([]model.UserResponseDTO, error) {
	newCtx, span := s.tracer.Start(ctx, "GetUsers.Service")
	defer span.End()

	time.Sleep(time.Microsecond * 10)

	users, err := s.repo.GetUsers(newCtx)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOpFailed
		}
	}

	dtos := make([]model.UserResponseDTO, len(users))

	for index, value := range users {
		dtos[index] = ConvertUserToUserReponseDTO(&value)
	}

	return users, nil
}

func (s *LocalUserService) GetById(ctx context.Context, id int) (*model.UserResponseDTO, error) {
	ctx, span := s.tracer.Start(ctx, "GetById.Service")
	defer span.End()

	user, err := s.getUserFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		zap.L().Debug("Cache miss", zap.Int("id", id))

		user, err = s.repo.GetById(ctx, id) // Get from db in case of case miss
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOpFailed
		}
		go s.addToCache(user) // Add to cache on a separate goroutine asynchronously
	}

	userResponse := ConvertUserToUserReponseDTO(user)
	return &userResponse, nil
}

func (s *LocalUserService) InsertUser(ctx context.Context, user model.UserCreateDTO) error {
	ctx, span := s.tracer.Start(ctx, "InsertUser.InsertUser")
	defer span.End()

	newuser := model.User{
		Id:      s.repo.Count() + 1,
		Name:    user.Name,
		City:    user.City,
		State:   user.State,
		Country: user.Country,
	}
	err := s.repo.InsertUser(ctx, newuser)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalUserService) UpdateUser(ctx context.Context, id int, new model.UserUpdateDTO) error {
	ctx, span := s.tracer.Start(ctx, "UpdateUser.UpdateUser")
	defer span.End()

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

	err := s.repo.UpdateUser(ctx, id, updateduser)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalUserService) DeleteUser(ctx context.Context, id int) error {
	ctx, span := s.tracer.Start(ctx, "DeleteUser.Service")
	defer span.End()

	err := s.repo.DeleteUser(ctx, id)
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
