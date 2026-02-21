package controller

import (
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
	"go.uber.org/zap"
)

type UsersController struct {
	userservice    service.UserService
	postservice    service.PostsService
	commentservice service.CommentService

	logger *zap.Logger
}

func NewUsersController(userService service.UserService, postService service.PostsService, commentService service.CommentService, logger *zap.Logger) *UsersController {
	return &UsersController{
		userservice:    userService,
		postservice:    postService,
		commentservice: commentService,
		logger:         logger,
	}
}

func (c *UsersController) GetUsers(w http.ResponseWriter, r *http.Request) {
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
		SendProblemDetails(w, "ValidationError", []model.ProblemDetailsError{
			{
				Field:   "limit",
				Message: "Provided limit is out of range. Valid: 1-10",
				Code:    "PARAMETER_MALFORMED",
			},
		}, r.URL.String())
		return
	}

	users, _ := c.userservice.GetUsers() // No point of error handling here as empty row will return [] and 200 status
	paginatedResponse := Paginate(users, cursor, limit, "users", "http://localhost")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(paginatedResponse)
}

func (c *UsersController) GetById(w http.ResponseWriter, r *http.Request) {
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

	user, err := c.userservice.GetById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRecord) {
			SendProblemDetails(w, "NotFound", nil, r.URL.String())
			return
		}
	}
	json.NewEncoder(w).Encode(user)
}

// GET /users/xxx-xxx-xxx/posts?limit=x
func (c *UsersController) GetPostsOfUser(w http.ResponseWriter, r *http.Request) {
	userString := r.PathValue("id")
	userId, err := strconv.Atoi(userString)
	if err != nil {
		http.Error(w, "Malformed id string", http.StatusBadRequest)
	}

	postResponse, err := c.postservice.GetPostsOfUser(userId)
	paginatedResponse := Paginate(postResponse, "", 10, "users", "http://localhost:9000")
	json.NewEncoder(w).Encode(paginatedResponse)
}

func (c *UsersController) PostUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	user := model.UserCreateDTO{}
	json.NewDecoder(r.Body).Decode(&user)
	err := c.userservice.InsertUser(user)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *UsersController) PatchUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	// testing only
	patch := model.UserUpdateDTO{}
	json.NewDecoder(r.Body).Decode(&patch)
	err := c.userservice.UpdateUser(patch.Id, patch)
	if err != nil {
		SendProblemDetails(w, "Error", nil, r.URL.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
}

func (c *UsersController) PutUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	w.WriteHeader(http.StatusNoContent)

	w.Write([]byte("Users Put route"))
}

func (c *UsersController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

	deleteuser := model.UserDeleteDTO{}
	json.NewDecoder(r.Body).Decode(&deleteuser)
	err := c.userservice.DeleteUser(deleteuser.Id)
	if err != nil {
		SendProblemDetails(w, "NotFound", nil, r.URL.String())
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

func SendProblemDetails(w http.ResponseWriter, problemType string, errors []model.ProblemDetailsError, route string) {
	switch problemType {
	case "NotFound":
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/error-not-found", "Resource not found", "The requested resource was not found", route, errors, http.StatusNotFound)
		}
	case "ValidationError":
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/malformed-parameter", "Validation error", "There's validation error", route, errors, http.StatusBadRequest)
		}
	case "Error":
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/internal-error", "Internal error", "The requested operation failed due to internal server error", route, errors, http.StatusInternalServerError)
		}
	case "Forbidden":
		{
			SendProblemDetailsCustom(w, "https://api.example.com/docs/forbidden", "Access denied", "Access to requested resource denied", route, errors, http.StatusForbidden)
		}
	case "Unauthorised":
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
