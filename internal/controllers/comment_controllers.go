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

type CommentsController struct {
	userservice    service.UserService
	postservice    service.PostsService
	commentservice service.CommentService
	logger         *zap.Logger
}

func NewCommentsController(userService service.UserService, postService service.PostsService, commentService service.CommentService, logger *zap.Logger) *CommentsController {
	return &CommentsController{
		userservice:    userService,
		postservice:    postService,
		commentservice: commentService,
		logger:         logger,
	}
}

func (c *CommentsController) GetComments(w http.ResponseWriter, r *http.Request) {
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
	if limit < 1 || limit > 10 {
		// limit = 10
		SendProblemDetails(w, "ValidationError", []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit is out of range. Valid: 1-10",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	comments, _ := c.commentservice.GetComments() // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(comments, cursor, limit, "posts", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *CommentsController) GetById(w http.ResponseWriter, r *http.Request) {
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

	comment, err := c.commentservice.GetById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, "NotFound", nil, r.URL.String())
			return
		}
	}
	json.NewEncoder(w).Encode(comment)
}

func (c *CommentsController) PostComment(w http.ResponseWriter, r *http.Request) {
	incoming := model.CommentCreateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.InsertComment(incoming)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *CommentsController) PatchComment(w http.ResponseWriter, r *http.Request) {
	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.UpdateComment(incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsController) PutComment(w http.ResponseWriter, r *http.Request) {
	incoming := model.CommentUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.UpdateComment(incoming.Id, incoming)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsController) DeleteComment(w http.ResponseWriter, r *http.Request) {
	incoming := model.CommentDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&incoming)
	err := c.commentservice.DeleteComment(incoming.Id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, "NotFound", nil, r.URL.String())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
