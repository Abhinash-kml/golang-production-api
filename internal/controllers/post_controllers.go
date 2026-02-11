package controller

import (
	"encoding/json"
	"net/http"

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

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	posts, _ := c.service.GetPosts()
	encoder.Encode(posts)
}

func (c *PostsController) PostPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	// c.service.InsertPost(model.Post{Id: })

	w.Write([]byte("Posts POST route"))
}

func (c *PostsController) PutPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	w.Write([]byte("Posts"))
}

func (c *PostsController) PatchPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	w.Write([]byte("Posts PATCH route"))
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

}
