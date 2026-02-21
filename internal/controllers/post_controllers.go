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
	cursor := r.URL.Query().Get("cursor")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		SendProblemDetails(w, "ValidationError", []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit cannot be converted to internal representation",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}
	if limit < 1 || limit > 100 {
		SendProblemDetails(w, "ValidationError", []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit is out of range. Valid: 1-10",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
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
		SendProblemDetails(w, "ValidationError", []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id is malformed",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	post, err := c.postservice.GetById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, "NotFound", nil, r.URL.String())
			return
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
		SendProblemDetails(w, "ValidationError", []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id is malformed",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	commentResponse, err := c.commentservice.GetCommentsOfPost(postId)
	paginatedResponse := Paginate(commentResponse, "", 10, "users", "http://localhost:9000")
	json.NewEncoder(w).Encode(paginatedResponse)
}

func (c *PostsController) PostPost(w http.ResponseWriter, r *http.Request) {
	incoming := model.PostCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.InsertPost(incoming)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *PostsController) PutPost(w http.ResponseWriter, r *http.Request) {
	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.UpdatePost(incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) PatchPost(w http.ResponseWriter, r *http.Request) {
	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.UpdatePost(incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	incoming := model.PostDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.DeletePost(incoming.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, "NotFound", nil, r.URL.String())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
