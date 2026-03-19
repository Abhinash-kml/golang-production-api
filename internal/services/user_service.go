package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	UpdateUser(context.Context, model.UserUpdateDTO) error
	ReplaceUser(context.Context, model.UserReplaceDTO) error
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
	ctx, span := s.tracer.Start(ctx, "GetUsers.Service")
	defer span.End()

	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error getting users from repository")
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoRecord
		} else {
			return nil, err
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

	span.SetAttributes(attribute.Int("user.id", id))

	user, err := s.getUserFromCache(id)
	if err != nil && errors.Is(err, redis.Nil) {
		span.RecordError(err)
		zap.L().Debug("Cache miss", zap.Int("id", id))

		user, err = s.repo.GetById(ctx, id) // Get from db in case of case miss
		if errors.Is(err, repository.ErrNoRecord) || errors.Is(err, sql.ErrNoRows) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to fetch user from repository")
			return nil, repository.ErrNoRecord
		}
		go s.addToCache(user) // Add to cache on a separate goroutine asynchronously
	}

	userResponse := ConvertUserToUserReponseDTO(user)
	return &userResponse, nil
}

func (s *LocalUserService) InsertUser(ctx context.Context, user model.UserCreateDTO) error {
	ctx, span := s.tracer.Start(ctx, "InsertUser.InsertUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.name", user.Name),
		attribute.String("user.city", user.City),
		attribute.String("user.state", user.State),
		attribute.String("user.country", user.Country))

	newuser := model.User{
		Id:      s.repo.Count() + 1,
		Name:    user.Name,
		City:    user.City,
		State:   user.State,
		Country: user.Country,
	}
	err := s.repo.InsertUser(ctx, newuser)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error inserting user in repository")
		return err
	}

	return nil
}

func (s *LocalUserService) UpdateUser(ctx context.Context, dto model.UserUpdateDTO) error {
	ctx, span := s.tracer.Start(ctx, "UpdateUser.UpdateUser")
	defer span.End()

	patches, _ := jsonpatch.DecodePatch(dto.Patch)

	span.SetAttributes(attribute.Int("user.id", dto.Id),
		attribute.Int("user.patches", len(patches)))

	err := s.repo.UpdateUser(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error upadating user in repository")
		return err
	}

	return nil
}

func (e *LocalUserService) ReplaceUser(ctx context.Context, dto model.UserReplaceDTO) error {
	ctx, span := e.tracer.Start(ctx, "ReplaceUser.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", dto.Id),
		attribute.String("user.name", dto.Name),
		attribute.String("user.city", dto.City),
		attribute.String("user.state", dto.State),
		attribute.String("user.country", dto.Country))

	err := e.repo.ReplaceUser(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error replacing user in repository")
		return err
	}
	return nil
}

func (s *LocalUserService) DeleteUser(ctx context.Context, id int) error {
	ctx, span := s.tracer.Start(ctx, "DeleteUser.Service")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", id))

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error deleting user in repository")
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

func HandleRepositoryError(span trace.Span, err error, resource string) {
	span.RecordError(err)
	if errors.Is(err, repository.ErrNoRecord) {
		span.SetAttributes(attribute.Bool(resource+".found", false))
	} else {
		span.SetStatus(codes.Error, "internal server error")
	}
}
