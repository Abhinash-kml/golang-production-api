package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type PostsController struct {
	userservice    service.UserService
	postservice    service.PostsService
	commentservice service.CommentService
	logger         *zap.Logger
	tracer         oteltracer.Tracer
}

func NewPostsController(userService service.UserService, postService service.PostsService, commentService service.CommentService, logger *zap.Logger, tracer oteltracer.Tracer) *PostsController {
	return &PostsController{
		userservice:    userService,
		postservice:    postService,
		commentservice: commentService,
		logger:         logger,
		tracer:         tracer,
	}
}

func (c *PostsController) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "GetPosts.Controller")
	defer span.End()

	cursor := r.URL.Query().Get("cursor")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit cannot be converted to internal representation",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}
	if limit < 1 || limit > 100 {
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit is out of range. Valid: 1-10",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	posts, _ := c.postservice.GetPosts(ctx) // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(posts, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *PostsController) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "GetById.Controller")
	defer span.End()

	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id is malformed",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	post, err := c.postservice.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, ProblemNotFound, nil, r.URL.String())
			return
		}
	}
	json.NewEncoder(w).Encode(post)
}

// Should this belong in posts controller or comments controller file ?
// GET posts/xxx-xxx-xxx/comments?limit=x
func (c *PostsController) GetCommentsOfPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "GetCommentsOfPost.Controller")
	defer span.End()

	postIdString := r.PathValue("id")
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id is malformed",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	commentResponse, err := c.commentservice.GetCommentsOfPost(ctx, postId)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}
	paginatedResponse := Paginate(commentResponse, "", 10, "users", "http://localhost:9000")
	json.NewEncoder(w).Encode(paginatedResponse)
}

func (c *PostsController) PostPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PostPost.Controller")
	defer span.End()

	incoming := model.PostCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.InsertPost(ctx, incoming)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *PostsController) PutPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PutPost.Controller")
	defer span.End()

	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.UpdatePost(ctx, incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) PatchPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PatchPost.Controller")
	defer span.End()

	incoming := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.UpdatePost(ctx, incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "DeletePost.Controller")
	defer span.End()

	incoming := model.PostDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.postservice.DeletePost(ctx, incoming.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, ProblemNotFound, nil, r.URL.String())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
