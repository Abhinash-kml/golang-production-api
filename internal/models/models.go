package model

import "time"

type User struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type UserRequest struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Aliased to main type
type UserResponse = User

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
	Comments  []Comment `json:"comments"`
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

type PostRequest struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type PostResponse struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
}

type Comment struct {
	Id          int    `json:"id"`
	CommenterId int    `json:"commenterid"`
	PostId      int    `json:"postid"`
	Body        string `json:"body"`
	Likes       int    `json:"likes"`
}

type CommentRequest struct {
	Id int `json:"id"`
}

type CommentResponse struct {
	Id          int    `json:"id"`
	CommenterId int    `json:"commenterid"`
	Body        string `json:"body"`
	Likes       int    `json:"likes"`
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
