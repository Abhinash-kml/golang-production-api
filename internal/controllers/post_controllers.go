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
		span.RecordError(err)
		span.SetStatus(codes.Error, "error converting limit to integer")
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

	posts, _ := c.postservice.GetPosts(ctx) // No point of error handling here as empty row will return [] and 200 status
	if len(posts) != 0 {
		span.SetAttributes(attribute.Bool("posts.found", true), attribute.Int("posts.num", len(posts)))
	} else {
		span.SetAttributes(attribute.Bool("posts.found", false))
	}
	paginatedResponse := Paginate(posts, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(paginatedResponse); err != nil {
		span.RecordError(err)
	}
}

func (c *PostsController) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "GetById.Controller")
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

	span.SetAttributes(attribute.Int("post.id", id))

	post, err := c.postservice.GetById(ctx, id)
	if err != nil {
		HandleServiceError(w, r, span, err, "post")
		return
	}

	if err := json.NewEncoder(w).Encode(post); err != nil {
		span.RecordError(err)
	}
}

// Should this belong in posts controller or comments controller file ?
// GET posts/xxx-xxx-xxx/comments?limit=x
func (c *PostsController) GetCommentsOfPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "GetCommentsOfPost.Controller")
	defer span.End()

	postIdString := r.PathValue("id")
	postId, err := strconv.Atoi(postIdString)
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

	span.SetAttributes(attribute.Int("post.id", postId))

	commentResponse, err := c.commentservice.GetCommentsOfPost(ctx, postId)
	if err != nil {
		HandleServiceError(w, r, span, err, "comments-of-post")
		return
	}

	paginatedResponse := Paginate(commentResponse, "", 10, "users", "http://localhost:9000")
	if err := json.NewEncoder(w).Encode(paginatedResponse); err != nil {
		span.RecordError(err)
	}
}

func (c *PostsController) PostPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PostPost.Controller")
	defer span.End()

	incoming := model.PostCreateDTO{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to json decode postcreatedto")
		return
	}

	span.SetAttributes(attribute.Int("post.authorid", incoming.AuthorID),
		attribute.String("post.title", incoming.Title),
		attribute.String("post.body", incoming.Body))

	err := c.postservice.InsertPost(ctx, incoming)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error inserting new post")
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TODO: Add span attributes as per json merge patch
func (c *PostsController) PutPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PutPost.Controller")
	defer span.End()

	dto := model.PostReplaceDTO{}
	json.NewDecoder(r.Body).Decode(&dto)

	span.SetAttributes(attribute.Int("post.id", dto.Id),
		attribute.String("post.title", dto.Title),
		attribute.String("post.body", dto.Body),
		attribute.Int("post.likes", dto.Likes))

	err := c.postservice.ReplacePost(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error replacing post")
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TODO: Add span attributes as per json merge patch
func (c *PostsController) PatchPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PatchPost.Controller")
	defer span.End()

	dto := model.PostUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&dto)

	span.SetAttributes(attribute.Int("post.id", dto.Id),
		attribute.Int("post.patch.num", len(dto.Patches)))

	err := c.postservice.UpdatePost(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error updating post")
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *PostsController) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "DeletePost.Controller")
	defer span.End()

	incoming := model.PostDeleteDTO{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to json decode postdeletedto")
		return
	}

	span.SetAttributes(attribute.Int("post.id", incoming.Id))

	err := c.postservice.DeletePost(ctx, incoming.Id)
	if err != nil {
		HandleServiceError(w, r, span, err, "post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
