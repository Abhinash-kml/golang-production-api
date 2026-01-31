package controller

import (
	"encoding/json"
	"net/http"

	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.uber.org/zap"
)

type UsersController struct {
	service service.UserService
	logger  *zap.Logger
}

func NewUsersController(service service.UserService, logger *zap.Logger) *UsersController {
	return &UsersController{
		service: service,
		logger:  logger,
	}
}

func (c *UsersController) GetUsers(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	users, _ := c.service.GetUsers()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(users)
}

func (c *UsersController) PostUsers(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Users post route"))
}

func (c *UsersController) PatchUsers(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Users Patch route"))
}

func (c *UsersController) PutUsers(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Users Put route"))
}
