package controller

import (
	"encoding/json"
	"net/http"

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

	comments, _ := c.service.GetComments()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(comments)
}

func (c *CommentsController) PostComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	c.service.InsertComment(incoming)

	w.Write([]byte("OK"))
}

func (c *CommentsController) PatchComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	c.service.UpdateComment(incoming.Id, incoming)

	w.Write([]byte("OK"))
}

func (c *CommentsController) PutComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	c.service.UpdateComment(incoming.Id, incoming)

	w.Write([]byte("OK"))
}
