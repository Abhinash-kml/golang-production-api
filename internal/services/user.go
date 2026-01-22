package service

import model "github.com/abhinash-kml/go-api-server/internal/models"

type UserService interface {
	GetUsers() ([]model.User, error)
	InsertUsers([]model.User) error
	UpdateUsers([]model.User) error
	DeleteUsers([]model.User) error
}
