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

type CommentsController struct {
	userservice    service.UserService
	postservice    service.PostsService
	commentservice service.CommentService
	logger         *zap.Logger
	tracer         oteltracer.Tracer
}

func NewCommentsController(userService service.UserService, postService service.PostsService, commentService service.CommentService, logger *zap.Logger, tracer oteltracer.Tracer) *CommentsController {
	return &CommentsController{
		userservice:    userService,
		postservice:    postService,
		commentservice: commentService,
		logger:         logger,
		tracer:         tracer,
	}
}

func (c *CommentsController) GetComments(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "GetComments.Controller")
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
	if limit < 1 || limit > 10 {
		// limit = 10
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit is out of range. Valid: 1-10",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	comments, _ := c.commentservice.GetComments(ctx) // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(comments, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *CommentsController) GetById(w http.ResponseWriter, r *http.Request) {
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

	comment, err := c.commentservice.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, ProblemNotFound, nil, r.URL.String())
			return
		}
	}
	json.NewEncoder(w).Encode(comment)
}

func (c *CommentsController) PostComment(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PostComment.Controller")
	defer span.End()

	incoming := model.CommentCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.InsertComment(ctx, incoming)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *CommentsController) PatchComment(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PatchComment.Controller")
	defer span.End()

	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.UpdateComment(ctx, incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsController) PutComment(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PutComment.Controller")
	defer span.End()

	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.UpdateComment(ctx, incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsController) DeleteComment(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "DeleteComment.Controller")
	defer span.End()

	incoming := model.CommentDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.DeleteComment(ctx, incoming.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, ProblemError, nil, r.URL.String())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
