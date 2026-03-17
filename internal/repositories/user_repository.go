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
	jsonpatch "github.com/evanphx/json-patch/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
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
	UpdateUser(context.Context, int, model.UserUpdateDTO) error
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

	for index := range e.users {
		if e.users[index].Id == id {
			zap.L().Info("Repository", zap.Any("address", &e.users[index]))
			return &e.users[index], nil
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
func (e *InMemoryUsersRepository) UpdateUser(ctx context.Context, id int, user model.UserUpdateDTO) error {
	ctx, span := e.tracer.Start(ctx, "UpdateUser.Repository")
	defer span.End()

	patches, err := jsonpatch.DecodePatch(user.Patch)
	if err != nil {
		return errors.New("Failed to decode json patch")
	}

	var updatedUser *model.User

	for index := range e.users {
		if e.users[index].Id == id {
			updatedUser = &e.users[index]
			break
		}
	}

	for index := range patches {
		currentpatch := patches[index]
		op := currentpatch.Kind()
		fmt.Println("kind:", op)
		switch op {
		case "add":
		case "remove":
		case "replace":
			what, _ := currentpatch.Path()
			fmt.Println("path:", what)
			switch what {
			case "/name":
				iface, _ := currentpatch.ValueInterface()
				val, _ := iface.(string)
				updatedUser.Name = val
			case "/city":
				iface, _ := currentpatch.ValueInterface()
				val, _ := iface.(string)
				fmt.Println("Value:", val)
				updatedUser.City = val
			case "/state":
				iface, _ := currentpatch.ValueInterface()
				val, _ := iface.(string)
				updatedUser.State = val
			case "/country":
				iface, _ := currentpatch.ValueInterface()
				val, _ := iface.(string)
				updatedUser.Country = val
			}
		}
	}

	for index := range e.users {
		if e.users[index].Id == id {
			//e.users[index] = updatedUser
			fmt.Println(e.users[index])
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
