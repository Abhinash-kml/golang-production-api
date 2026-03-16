package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	ctx, span := c.tracer.Start(r.Context(), "GetComments.Controller")
	defer span.End()

	cursor := r.URL.Query().Get("cursor")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert provided limit to integer")
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
		span.SetStatus(codes.Error, "provided limit is out of range")
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit is out of range. Valid: 1-10",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	span.SetAttributes(attribute.String("cursor", cursor), attribute.Int("limit", limit))

	comments, _ := c.commentservice.GetComments(ctx) // No point of error handling here as empty row will return [] and 200 status
	if len(comments) != 0 {
		span.SetAttributes(attribute.Bool("comments.found", true), attribute.Int("comments.num", len(comments)))
	} else {
		span.SetAttributes(attribute.Bool("comments.found", true))
	}

	paginatedResponse := Paginate(comments, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(paginatedResponse); err != nil {
		span.RecordError(err)
	}
}

func (c *CommentsController) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "GetById.Controller")
	defer span.End()

	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert provided id to integer")
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id is malformed",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	span.SetAttributes(attribute.Int("comment.id", id))

	comment, err := c.commentservice.GetById(ctx, id)
	if err != nil {
		HandleServiceError(w, r, span, err, "comment")
		return
	}

	if err := json.NewEncoder(w).Encode(comment); err != nil {
		span.RecordError(err)
	}
}

func (c *CommentsController) PostComment(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "PostComment.Controller")
	defer span.End()

	incoming := model.CommentCreateDTO{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to json decode commentcreatedto")
		return
	}

	span.SetAttributes(attribute.Int("comment.authorid", incoming.Authorid),
		attribute.Int("comment.postid", incoming.Postid),
		attribute.String("comment.body", incoming.Body))

	err := c.commentservice.InsertComment(ctx, incoming)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error inserting new comment")
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TODO: Implement spans as per json merge patch
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

// TODO: Implement spans as per json merge patch
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
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error json decoding commentdeletedto")
		return
	}

	span.SetAttributes(attribute.Int("comment.id", incoming.Id))

	err := c.commentservice.DeleteComment(ctx, incoming.Id)
	if err != nil {
		HandleServiceError(w, r, span, err, "comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
