package repository

import (
	"errors"
	"slices"

	model "github.com/abhinash-kml/go-api-server/internal/models"
)

var (
	ErrSetupFailed     = errors.New("Repository setup failed")
	ErrNoUsers         = errors.New("No user in repository")
	ErrUndefinedUsers  = errors.New("Undefined users")
	ErrZeroLengthSlice = errors.New("Provided slice is of zero length")
	ErrUserDoesntExist = errors.New("Provided users doesn't exist")
)

type UserRepository interface {
	// Initialize
	Setup() error

	// CRUD logics
	GetUsers() ([]model.User, error)
	InsertUsers([]model.User) error
	UpdateUsers([]model.User, []model.User) error
	DeleteUsers([]model.User) error
}

type InMemoryRepository struct {
	users []model.User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (e *InMemoryRepository) Setup() error {
	users := []model.User{
		{
			Id:      1,
			Name:    "Neo",
			City:    "Kolkata",
			State:   "West Bengal",
			Country: "India",
		},
		{
			Id:      2,
			Name:    "Abhinash",
			City:    "Kolkata",
			State:   "West Bengal",
			Country: "India",
		},
		{
			Id:      3,
			Name:    "Komal",
			City:    "Kolkata",
			State:   "West Bengal",
			Country: "India",
		},
		{
			Id:      4,
			Name:    "Riya",
			City:    "Ranchi",
			State:   "Jharkhand",
			Country: "India",
		},
		{
			Id:      5,
			Name:    "Jyotika",
			City:    "Kolkata",
			State:   "West Bengal",
			Country: "India",
		},
	}

	for _, value := range users {
		e.users = append(e.users, value)
	}

	return nil
}

func (e *InMemoryRepository) GetUsers() ([]model.User, error) {
	if len(e.users) <= 0 {
		return nil, ErrNoUsers
	}

	return e.users, nil
}

func (e *InMemoryRepository) InsertUsers(users []model.User) error {
	if len(users) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, value := range users {
		e.users = append(e.users, value)
	}

	return nil
}

func (e *InMemoryRepository) UpdateUsers(old, new []model.User) error {
	if len(old) <= 0 || len(new) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, value := range e.users {
		for i := 0; i < len(old); i++ {
			if value.Id == old[i].Id {
				value = new[i]
			}
		}
	}

	return nil
}

func (e *InMemoryRepository) DeleteUsers(users []model.User) error {
	if len(users) <= 0 {
		return ErrZeroLengthSlice
	}

	e.users = slices.DeleteFunc(e.users, func(u model.User) bool {
		for _, value := range users {
			if u.Id == value.Id {
				return true
			}
		}

		return false
	})

	return nil
}
