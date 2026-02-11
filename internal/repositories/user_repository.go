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
	InsertUser(model.User) error
	UpdateUser(int, model.User) error
	DeleteUser(int) error
	Count() int
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

func (e *InMemoryUsersRepository) InsertUser(user model.User) error {
	e.users = append(e.users, user)

	return nil
}

func (e *InMemoryUsersRepository) UpdateUser(id int, user model.User) error {
	for index := range e.users {
		if e.users[index].Id == id {
			e.users[index] = user
			break
		}
	}

	return nil
}

func (e *InMemoryUsersRepository) DeleteUser(id int) error {

	users := slices.DeleteFunc(e.users, func(u model.User) bool {
		if u.Id == id {
			return true
		}

		return false
	})

	e.users = users
	return nil
}

func (e *InMemoryUsersRepository) Count() int {
	return len(e.users)
}
