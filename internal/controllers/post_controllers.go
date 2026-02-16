package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.uber.org/zap"
)

type PostsController struct {
	userservice    service.UserService
	postservice    service.PostsService
	commentservice service.CommentService
	logger         *zap.Logger
}

func NewPostsController(userService service.UserService, postService service.PostsService, commentService service.CommentService, logger *zap.Logger) *PostsController {
	return &PostsController{
		userservice:    userService,
		postservice:    postService,
		commentservice: commentService,
		logger:         logger,
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

	posts, _ := c.postservice.GetPosts() // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(posts, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *PostsController) GetById(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Malformed id string", http.StatusBadRequest)
	}

	post, err := c.postservice.GetById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			http.Error(w, "No Record", http.StatusNotFound)
		}
	}
	json.NewEncoder(w).Encode(post)
}

// Should this belong in posts controller or comments controller file ?
// GET posts/xxx-xxx-xxx/comments?limit=x
func (c *PostsController) GetCommentsOfPost(w http.ResponseWriter, r *http.Request) {
	postIdString := r.PathValue("id")
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		http.Error(w, "Malformed id string", http.StatusBadRequest)
	}

	commentResponse, err := c.commentservice.GetCommentsOfPost(postId)
	paginatedResponse := Paginate(commentResponse, "", 10, "users", "http://localhost:9000")
	json.NewEncoder(w).Encode(paginatedResponse)
}

func (c *PostsController) PostPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.InsertPost(incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *PostsController) PutPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.UpdatePost(incoming.Id, incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) PatchPost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.UpdatePost(incoming.Id, incoming)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	incoming := model.PostDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.DeletePost(incoming.Id)
	if err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
