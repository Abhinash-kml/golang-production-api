package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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

type InMemoryUsersRepository struct {
	users []model.User
}

func NewInMemoryUsersRepository() *InMemoryUsersRepository {
	return &InMemoryUsersRepository{}
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

func (e *InMemoryUsersRepository) GetUsers() ([]model.User, error) {
	if len(e.users) <= 0 {
		return nil, ErrNoUsers
	}

	return e.users, nil
}

func (e *InMemoryUsersRepository) InsertUsers(users []model.User) error {
	if len(users) <= 0 {
		return ErrZeroLengthSlice
	}

	for _, value := range users {
		e.users = append(e.users, value)
	}

	return nil
}

func (e *InMemoryUsersRepository) UpdateUsers(old, new []model.User) error {
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

func (e *InMemoryUsersRepository) DeleteUsers(users []model.User) error {
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
