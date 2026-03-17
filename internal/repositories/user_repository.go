package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltracer "go.opentelemetry.io/otel/trace"
)

var (
	ErrSetupFailed     = errors.New("Repository setup failed")
	ErrZeroLengthSlice = errors.New("Provided slice is of zero length")
	ErrNoRecord        = errors.New("No record in database")
)

type UserRepository interface {
	// Initialize
	Setup() error

	// CRUD logics
	GetUsers(context.Context) ([]model.User, error)
	GetById(context.Context, int) (*model.User, error)
	InsertUser(context.Context, model.User) error
	UpdateUser(context.Context, int, model.User) error
	DeleteUser(context.Context, int) error
	Count() int
}

type InMemoryUsersRepository struct {
	users  []model.User
	tracer oteltracer.Tracer
}

func NewInMemoryUsersRepository(tracer oteltracer.Tracer) *InMemoryUsersRepository {
	return &InMemoryUsersRepository{tracer: tracer}
}

func (e *InMemoryUsersRepository) Setup() error {
	users := make([]model.User, 0, 150)

	file, err := os.OpenFile("./mocks/users.json", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal("Error opening users mock data file. Error:", err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	if err != nil {
		file.Close()
		log.Fatal("Failed to decode, Error: ", err.Error())
	}

	fmt.Println("Succesfully read", len(users), "users from mocks")
	e.users = users

	return nil
}

func (e *InMemoryUsersRepository) GetUsers(ctx context.Context) ([]model.User, error) {
	_, span := e.tracer.Start(ctx, "GetUsers.Repository")
	defer span.End()

	if len(e.users) <= 0 {
		span.SetAttributes(attribute.Bool("users.found", true))
		span.SetStatus(codes.Error, "failed to fetch users in repository")
		return nil, ErrNoRecord
	}

	span.SetAttributes(attribute.Bool("users.found", true), attribute.Int("users.num", len(e.users)))

	return e.users, nil
}

func (e *InMemoryUsersRepository) GetById(ctx context.Context, id int) (*model.User, error) {
	ctx, span := e.tracer.Start(ctx, "GetById.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", id))

	for _, value := range e.users {
		if value.Id == id {
			return &value, nil
		}
	}

	span.SetAttributes(attribute.Bool("user.found", false))
	span.SetStatus(codes.Error, "failed to fetch user in repoitory")
	return nil, ErrNoRecord
}

func (e *InMemoryUsersRepository) InsertUser(ctx context.Context, user model.User) error {
	ctx, span := e.tracer.Start(ctx, "InsertUser.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", user.Id),
		attribute.String("user.name", user.Name),
		attribute.String("user.city", user.City),
		attribute.String("user.state", user.State),
		attribute.String("user.country", user.Country))

	e.users = append(e.users, user)

	return nil
}

// TODO: Implement as per JSON Merge Patch
func (e *InMemoryUsersRepository) UpdateUser(ctx context.Context, id int, user model.User) error {
	ctx, span := e.tracer.Start(ctx, "UpdateUser.Repository")
	defer span.End()

	for index := range e.users {
		if e.users[index].Id == id {
			e.users[index] = user
			break
		}
	}

	return nil
}

func (e *InMemoryUsersRepository) DeleteUser(ctx context.Context, id int) error {
	ctx, span := e.tracer.Start(ctx, "DeleteUser.Repository")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", id))

	oldlen := len(e.users)
	e.users = slices.DeleteFunc(e.users, func(u model.User) bool {
		if u.Id == id {
			return true
		}

		return false
	})
	newlen := len(e.users)

	if newlen != oldlen {
		span.SetAttributes(attribute.Bool("user.deleted", true))
	} else {
		span.SetAttributes(attribute.Bool("user.deleted", false))
	}
	return nil
}

func (e *InMemoryUsersRepository) Count() int {
	return len(e.users)
}
