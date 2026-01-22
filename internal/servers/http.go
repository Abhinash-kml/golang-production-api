package servers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	service "github.com/abhinash-kml/go-api-server/internal/services"
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

	// Services
	userservice service.UserService
	// postservice
	// commentservice
	// friendsservice

	// Repositories
	userrepository repository.UserRepository
	// postsrepository
	// commentsrepository
	// friendsrepository

	// Logger
	logger zap.Logger
}

func NewHttpServer(Addr string, ReadTimeOut, WriteTimeout, IdleTimeout time.Duration, MaxHeaderBytes int) *HttpServer {
	return &HttpServer{
		Addr:           Addr,
		ReadTimeout:    ReadTimeOut,
		WriteTimeout:   WriteTimeout,
		IdleTimeout:    IdleTimeout,
		MaxHeaderBytes: MaxHeaderBytes,
	}
}

func (s *HttpServer) SetupServices() error {
	return nil
}

func (s *HttpServer) SetupRepository() error {
	return nil
}

func (s *HttpServer) SetupRoutes() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", controller.GetUsers)
	mux.HandleFunc("POST /users", controller.PostUsers)
	mux.HandleFunc("PUT /users", controller.PutUsers)
	mux.HandleFunc("PATCH /users", controller.PatchUsers)

	s.Handler = mux
	return nil
}

func (s *HttpServer) Start() error {
	fmt.Println("Starting HTTP server on ", s.Addr)
	return http.ListenAndServe(s.Addr, s.Handler)
}

func (s *HttpServer) Stop() error {
	return nil
}
