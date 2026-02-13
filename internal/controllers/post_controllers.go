package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.uber.org/zap"
)

type PostsController struct {
	service service.PostsService
	logger  *zap.Logger
}

func NewPostsController(service service.PostsService, logger *zap.Logger) *PostsController {
	return &PostsController{
		service: service,
		logger:  logger,
	}
}

func (c *PostsController) GetPosts(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	cursor := r.URL.Query().Get("cursor")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Cannot convert provided limit to integer", http.StatusBadRequest)
	}
	if limit < 1 || limit > 100 {
		http.Error(w, "Malformed query limit. Correct range: 1-100", http.StatusBadRequest)
	}

	posts, _ := c.service.GetPosts() // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(posts, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *PostsController) PostPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.InsertPost(incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *PostsController) PutPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.UpdatePost(incoming.Id, incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) PatchPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.UpdatePost(incoming.Id, incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.DeletePost(incoming.Id)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
