package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.uber.org/zap"
)

type CommentsController struct {
	service service.CommentService
	logger  *zap.Logger
}

func NewCommentsController(service service.CommentService, logger *zap.Logger) *CommentsController {
	return &CommentsController{
		service: service,
		logger:  logger,
	}
}

func (c *CommentsController) GetComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	cursor := r.URL.Query().Get("cursor")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Cannot convert provided limit to integer", http.StatusBadRequest)
	}
	if limit < 1 || limit > 100 {
		http.Error(w, "Malformed query limit. Correct range: 1-100", http.StatusBadRequest)
	}

	comments, _ := c.service.GetComments() // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(comments, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *CommentsController) PostComment(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.InsertComment(incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *CommentsController) PatchComment(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.UpdateComment(incoming.Id, incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsController) PutComment(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.UpdateComment(incoming.Id, incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsController) DeleteComment(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.service.DeleteComment(incoming.Id)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
