package model

import "time"

type User struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
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
	Id      int    `json:"id"`
	What    string `json:"what"`
	NewData string `json:"newdata"`
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
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Likes     int       `json:"likes"`
	CreatorId int       `json:"creatorid"`
	CreatedAt time.Time `json:"createdat"`
}

func NewPost(id int, title, body string, likes, creatorid int, createdat time.Time) *Post {
	return &Post{
		Id:        id,
		Title:     title,
		Body:      body,
		Likes:     likes,
		CreatorId: creatorid,
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
	Authorid int    `json:"authorid"`
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
	Id          int    `json:"id"`
	CommenterId int    `json:"commenterid"`
	PostId      int    `json:"postid"`
	Body        string `json:"body"`
	Likes       int    `json:"likes"`
}

type CommentRequestDTO struct {
	Id int `json:"id"`
}

type CommentResponseDTO struct {
	Id          int    `json:"id"`
	CommenterId int    `json:"commenterid"`
	PostID      int    `json:"postid"`
	Body        string `json:"body"`
	Likes       int    `json:"likes"`
}

type CommentCreateDTO struct {
	Postid   int    `json:"postid"`
	Authorid int    `json:"authorid"`
	Body     string `json:"body"`
}

type CommentDeleteDTO struct {
	Id int `json:"id"`
}

type CommentUpdateDTO struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func NewComment(id, commenterid, postid int, body string, likes int) *Comment {
	return &Comment{
		Id:          id,
		CommenterId: commenterid,
		PostId:      postid,
		Body:        body,
		Likes:       likes,
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
