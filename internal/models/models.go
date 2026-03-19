package model

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Id      int    `json:"id" redis:"id"`
	Name    string `json:"name" redis:"name"`
	City    string `json:"city" redis:"city"`
	State   string `json:"state" redis:"state"`
	Country string `json:"country" redis:"country"`
}

type UserRequestDTO struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Aliased to main type
type UserResponseDTO = User

type UserCreateDTO struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type UserDeleteDTO struct {
	Id int `json:"id"`
}

type UserUpdateDTO struct {
	Id    int             `json:"id"`
	Patch json.RawMessage `json:"patch"`
}

type UserReplaceDTO struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

func NewUser(Id int, Name, City, State, Country string) *User {
	return &User{
		Id:      Id,
		Name:    Name,
		City:    City,
		State:   State,
		Country: Country,
	}
}

type Post struct {
	Id        int       `json:"id" redis:"id"`
	Title     string    `json:"title" redis:"title"`
	Body      string    `json:"body" redis:"body"`
	Likes     int       `json:"likes" redis:"likes"`
	AuthorID  int       `json:"author_id" redis:"author_id"`
	CreatedAt time.Time `json:"created_at" redis:"created_at"`
}

func NewPost(id int, title, body string, likes, authorid int, createdat time.Time) *Post {
	return &Post{
		Id:        id,
		Title:     title,
		Body:      body,
		Likes:     likes,
		AuthorID:  authorid,
		CreatedAt: createdat,
	}
}

type PostRequestDTO struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type PostResponseDTO struct {
	Id       int    `json:"id"`
	AuthorId int    `json:"authorid"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
}

type PostCreateDTO struct {
	AuthorID int    `json:"author_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

type PostDeleteDTO struct {
	Id int `json:"id"`
}

type PostUpdateDTO struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type PostReplaceDTO struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Likes int    `json:"likes"`
}

type Comment struct {
	Id       int    `json:"id" redis:"id"`
	AuthorID int    `json:"author_id" redis:"author_id"`
	PostId   int    `json:"postid" redis:"postid"`
	Body     string `json:"body" redis:"body"`
	Likes    int    `json:"likes" redis:"likes"`
}

type CommentRequestDTO struct {
	Id int `json:"id"`
}

type CommentResponseDTO struct {
	Id       int    `json:"id"`
	AuthorID int    `json:"author_id"`
	PostID   int    `json:"post_id"`
	Body     string `json:"body"`
	Likes    int    `json:"likes"`
}

type CommentCreateDTO struct {
	Postid   int    `json:"post_id"`
	Authorid int    `json:"author_id"`
	Body     string `json:"body"`
}

type CommentDeleteDTO struct {
	Id int `json:"id"`
}

type CommentUpdateDTO struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type CommentReplaceDTO struct {
	Id   int    `json:"id"`
	Body string `json:"string"`
}

func NewComment(id, authorid, postid int, body string, likes int) *Comment {
	return &Comment{
		Id:       id,
		AuthorID: authorid,
		PostId:   postid,
		Body:     body,
		Likes:    likes,
	}
}

// Pagination related types
type Links struct {
	Self     string `json:"self"`
	Previous string `json:"previous"`
	Next     string `json:"next"`
	First    string `json:"first"`
	Last     string `json:"last"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_page"`
}

type ApiPaginatedResponseDTO[T any] struct {
	Data  []T   `json:"data"`
	Links Links `json:"links"`
	Meta  Meta  `json:"meta"`
}

type ProblemDetailsResponse struct {
	Type     string                `json:"type"`
	Title    string                `json:"title"`
	Status   int                   `json:"status"`
	Detail   string                `json:"detail"`
	Instance string                `json:"instance"`
	Errors   []ProblemDetailsError `json:"errors,omitempty"`
}

type ProblemDetailsError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type AccessTokenRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type CustomJwtClaims struct {
	Role    string `json:"role"`
	Version string `json:"version"`
	jwt.RegisteredClaims
}
