package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	oteltracer "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type ProblemType int

const (
	ProblemNotFound ProblemType = iota
	ProblemValidationError
	ProblemError
	ProblemForbidden
	ProblemUnauthorized
)

type UsersController struct {
	userservice    service.UserService
	postservice    service.PostsService
	commentservice service.CommentService

	logger *zap.Logger
	tracer oteltracer.Tracer
}

func NewUsersController(userService service.UserService, postService service.PostsService, commentService service.CommentService, logger *zap.Logger, tracer oteltracer.Tracer) *UsersController {
	return &UsersController{
		userservice:    userService,
		postservice:    postService,
		commentservice: commentService,
		logger:         logger,
		tracer:         tracer,
	}
}

func (c *UsersController) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "GetUsers.Controller")
	defer span.End()

	cursor := r.URL.Query().Get("cursor")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to convert limit to numeric value")
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

	users, _ := c.userservice.GetUsers(ctx) // No point of error handling here as empty row will return [] and 200 status
	if len(users) != 0 {
		span.SetAttributes(attribute.Bool("users.found", true), attribute.Int("users.num", len(users)))
	} else {
		span.SetAttributes(attribute.Bool("users.found", false))
	}

	paginatedResponse := Paginate(users, cursor, limit, "users", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(paginatedResponse); err != nil {
		span.RecordError(err)
	}
}

func (c *UsersController) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "GetById.Controller")
	defer span.End()

	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		span.SetStatus(codes.Error, "malformed id")
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id is malformed",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	span.SetAttributes(attribute.Int("user.id", id))

	user, err := c.userservice.GetById(ctx, id)
	if err != nil {
		HandleServiceError(w, r, span, err, "user")
		return
	}

	span.SetAttributes(attribute.Bool("user.found", true))

	if err := json.NewEncoder(w).Encode(user); err != nil {
		span.RecordError(err)
	}
}

// GET /users/xxx-xxx-xxx/posts?limit=x
func (c *UsersController) GetPostsOfUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "GetPostsOfUser.Controller")
	defer span.End()

	userString := r.PathValue("id")
	userId, err := strconv.Atoi(userString)
	if err != nil {
		span.SetStatus(codes.Error, "malformed id")
		SendProblemDetails(w, ProblemValidationError, []model.ProblemDetailsError{
			{
				Field:   "id",
				Message: "Provided id string is malformed",
				Code:    "PARAMTER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	span.SetAttributes(attribute.Int("user.id", userId))

	postResponse, err := c.postservice.GetPostsOfUser(ctx, userId)
	if err != nil {
		HandleServiceError(w, r, span, err, "user-post")
		return
	}

	span.SetAttributes(attribute.Bool("user-post.found", true))

	paginatedResponse := Paginate(postResponse, "", 10, "users", "http://localhost:9000")
	if err := json.NewEncoder(w).Encode(paginatedResponse); err != nil {
		span.RecordError(err)
	}
}

func (c *UsersController) PostUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PostUser.Controller")
	defer span.End()

	user := model.UserCreateDTO{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "json decoding usercreatedto failed")
		return
	}

	// Validator
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(user)
	if err != nil {
		SendProblemDetails(w, ProblemValidationError, nil, r.URL.String())
		return
	}

	span.SetAttributes(attribute.String("user.name", user.Name),
		attribute.String("user.city", user.City),
		attribute.String("user.state", user.State),
		attribute.String("user.country", user.Country))

	err = c.userservice.InsertUser(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error inserting new user")
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TODO: Add spans as per json merge patch
func (c *UsersController) PatchUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PatchUser.Controller")
	defer span.End()

	dto := model.UserUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&dto)

	span.SetAttributes(attribute.Int("user.id", dto.Id),
		attribute.Int("user.patch.num", len(dto.Patches)))

	err := c.userservice.UpdateUser(ctx, dto)
	if err != nil {
		fmt.Println(err)
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *UsersController) PutUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "PutUser.Controller")
	defer span.End()

	var dto model.UserReplaceDTO
	json.NewDecoder(r.Body).Decode(&dto)
	err := c.userservice.ReplaceUser(ctx, dto)
	if err != nil {
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *UsersController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(context.Background(), "DeleteUser.Controller")
	defer span.End()

	deleteuser := model.UserDeleteDTO{}
	if err := json.NewDecoder(r.Body).Decode(&deleteuser); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "json decoding userdeletedto failed")
		return
	}

	span.SetAttributes(attribute.Int("user.id", deleteuser.Id))

	err := c.userservice.DeleteUser(ctx, deleteuser.Id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete user")
		SendProblemDetails(w, ProblemNotFound, nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Generic pagination helper to paginate incoming data from service layer
// TODO: Move to utils package
// TODO: Improve this
func Paginate[T any](data []T, currentCursorstring string, limit int, route, baseurl string) *model.ApiPaginatedResponseDTO[T] {
	// Extract pagination key from current cursor
	// If currunt cursor is nil / empty then present them first page
	var currentCursor int
	if currentCursorstring == "" {
		currentCursor = 0
	} else {
		bytes, err := base64.URLEncoding.DecodeString(currentCursorstring)
		if err != nil {
			log.Fatal("Failed to decode cursor from url")
		}

		currentCursor, err = strconv.Atoi(string(bytes))
		if err != nil {
			log.Fatal("Failed to convert decoded cursor to integer")
		}
	}

	encode := func(k int) string {
		return base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(k)))
	}

	dataLenth := len(data)
	var last int
	if currentCursor+limit > dataLenth {
		last = dataLenth
	} else {
		last = currentCursor + limit
	}

	pageSize := 10
	currentPage := currentCursor / pageSize
	totalPages := len(data) / pageSize
	nextPage := currentPage + 1
	prevPage := currentPage - 1
	if nextPage > totalPages {
		nextPage = 0
	}

	// Calculate Previous and Next Cursors
	selfCursor := currentCursor
	prevCursor := currentCursor - limit
	nextCursor := currentCursor + limit
	firstPageCursor := 0                                // Hardcoded, TODO: Maybe adapt to real data source
	lastpageCursor := ((totalPages + 1) - 1) * pageSize // +1 as pages are 0 based index

	calculateNextPageString := func(n int) string {
		if n <= 0 {
			return "null"
		}
		return fmt.Sprintf("%s/%s?cursor=%s&limit=%d", baseurl, route, encode(nextCursor), limit)
	}
	calculatePrevPageString := func(n int) string {
		if n < 0 {
			return "null"
		}
		return fmt.Sprintf("%s/%s?cursor=%s&limit=%d", baseurl, route, encode(prevCursor), limit)
	}

	response := &model.ApiPaginatedResponseDTO[T]{
		Data: data[currentCursor:last], // TODO: Fix this [overflow error]
		Links: model.Links{
			Self:     fmt.Sprintf("%s/%s?cursor=%s&limit=%d", baseurl, route, encode(selfCursor), limit),
			Previous: calculatePrevPageString(prevPage),
			Next:     calculateNextPageString(nextPage),
			First:    fmt.Sprintf("%s/%s?cursor=%s&limit=%d", baseurl, route, encode(firstPageCursor), limit),
			Last:     fmt.Sprintf("%s/%s?cursor=%s&limit=%d", baseurl, route, encode(lastpageCursor), limit),
		},
		Meta: model.Meta{
			CurrentPage: currentPage,
			TotalPages:  totalPages,
		},
	}

	return response
}

func SendProblemDetails(w http.ResponseWriter, ptype ProblemType, errors []model.ProblemDetailsError, route string) {
	switch ptype {
	case ProblemNotFound:
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/error-not-found", "Resource not found", "The requested resource was not found", route, errors, http.StatusNotFound)
		}
	case ProblemValidationError:
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/malformed-parameter", "Validation error", "There's validation error", route, errors, http.StatusBadRequest)
		}
	case ProblemError:
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/internal-error", "Internal server error", "The requested operation failed due to internal server error", route, errors, http.StatusInternalServerError)
		}
	case ProblemForbidden:
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/forbidden", "Access denied", "Access to requested resource denied", route, errors, http.StatusForbidden)
		}
	case ProblemUnauthorized:
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/unauthorized", "Unauthorized", "Authorization is required", route, errors, http.StatusUnauthorized)
		}
	}
}

func SendProblemDetailsCustom(w http.ResponseWriter, Type, title, details, instance string, errors []model.ProblemDetailsError, status int) {
	reponse := model.ProblemDetailsResponse{
		Type:     Type,
		Title:    title,
		Detail:   details,
		Instance: instance,
		Status:   status,
		Errors:   errors,
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(reponse)
}

func HandleServiceError(w http.ResponseWriter, r *http.Request, span trace.Span, err error, resource string) {
	span.RecordError(err)
	if errors.Is(err, repository.ErrNoRecord) {
		span.SetAttributes(attribute.Bool(resource+".found", false))
		SendProblemDetails(w, ProblemNotFound, nil, r.URL.String())
	} else {
		span.SetStatus(codes.Error, "internal server error")
		SendProblemDetails(w, ProblemError, nil, r.URL.String())
	}
}
