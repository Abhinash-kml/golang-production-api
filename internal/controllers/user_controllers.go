package controller

import (
	"encoding/json"
	"net/http"

	model "github.com/abhinash-kml/go-api-server/internal/models"
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
	// state := r.URL.Query().Get("state")

	users, _ := c.service.GetUsers() // No point of error handling here as empty row will return [] and 200 status
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(users)
}

func (c *UsersController) PostUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	user := model.UserCreateDTO{}
	json.NewDecoder(r.Body).Decode(&user)
	err := c.service.InsertUser(user)
	if err != nil {
		// TODO: Handle custom error here
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *UsersController) PatchUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	// testing only
	patch := model.UserUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&patch)
	err := c.service.UpdateUser(patch.Id, patch)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
}

func (c *UsersController) PutUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	w.WriteHeader(http.StatusNoContent)

	w.Write([]byte("Users Put route"))
}

func (c *UsersController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	deleteuser := model.UserDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&deleteuser)
	err := c.service.DeleteUser(deleteuser.Id)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
