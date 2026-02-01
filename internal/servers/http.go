package servers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	"go.uber.org/zap"
)

type HttpServer struct {
	Addr           string
	Handler        http.Handler
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	TLSConfig      *tls.Config

	// Controllers
	userscontroller controller.UsersController
	// postscontroller    controller.PostsController
	commentscontroller controller.CommentsController

	// Logger
	logger zap.Logger
}

func NewHttpServer(
	Addr string,
	ReadTimeOut,
	WriteTimeout,
	IdleTimeout time.Duration,
	MaxHeaderBytes int,
	UsersController controller.UsersController,
	CommentsController controller.CommentsController) *HttpServer {
	return &HttpServer{
		Addr:               Addr,
		ReadTimeout:        ReadTimeOut,
		WriteTimeout:       WriteTimeout,
		IdleTimeout:        IdleTimeout,
		MaxHeaderBytes:     MaxHeaderBytes,
		userscontroller:    UsersController,
		commentscontroller: CommentsController,
	}
}

func (s *HttpServer) SetupRoutes() error {
	mux := http.NewServeMux()

	// Users routes
	mux.HandleFunc("GET /users", s.userscontroller.GetUsers)
	mux.HandleFunc("POST /users", s.userscontroller.PostUsers)
	mux.HandleFunc("PUT /users", s.userscontroller.PutUsers)
	mux.HandleFunc("PATCH /users", s.userscontroller.PatchUsers)

	// Post routes
	// mux.HandleFunc("GET /posts", s.postscontroller.GetPosts)
	// mux.HandleFunc("POST /posts", s.postscontroller.PostPosts)
	// mux.HandleFunc("PUT /posts", s.postscontroller.PutPosts)
	// mux.HandleFunc("PATCH /posts", s.postscontroller.PatchPosts)

	// Comments routes
	mux.HandleFunc("GET /comments", s.commentscontroller.GetComments)
	mux.HandleFunc("POST /comments", s.commentscontroller.PostComments)
	mux.HandleFunc("PUT /comments", s.commentscontroller.PutComments)
	mux.HandleFunc("PATCH /comments", s.commentscontroller.PatchComments)

	s.Handler = mux
	return nil
}

func (s *HttpServer) Start() error {
	fmt.Println("Starting HTTP server on ", s.Addr)
	return http.ListenAndServe(s.Addr, s.Handler)
}

func (s *HttpServer) Stop() {
	fmt.Println("Shutting down http server...")
	os.Exit(0)
}
