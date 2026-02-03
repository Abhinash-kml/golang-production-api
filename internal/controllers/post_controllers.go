package controller

import (
	"encoding/json"
	"net/http"

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
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	posts, _ := c.service.GetPosts()
	encoder.Encode(posts)
}

func (c *PostsController) PostPosts(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Posts POST route"))
}

func (c *PostsController) PutPosts(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Posts"))
}

func (c *PostsController) PatchPosts(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection from", zap.String("IP", r.RemoteAddr))
	w.Write([]byte("Posts PATCH route"))
}
