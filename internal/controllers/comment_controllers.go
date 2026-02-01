package controller

import (
	"encoding/json"
	"net/http"

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
	comments, _ := c.service.GetComments()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(comments)
}

func (c *CommentsController) PostComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Post comments working"))
}

func (c *CommentsController) PatchComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Patch comments working"))
}

func (c *CommentsController) PutComments(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Put comments working"))
}
