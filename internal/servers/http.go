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

	// Logger
	logger zap.Logger
}

func NewHttpServer(Addr string, ReadTimeOut, WriteTimeout, IdleTimeout time.Duration, MaxHeaderBytes int, UsersController controller.UsersController) *HttpServer {
	return &HttpServer{
		Addr:            Addr,
		ReadTimeout:     ReadTimeOut,
		WriteTimeout:    WriteTimeout,
		IdleTimeout:     IdleTimeout,
		MaxHeaderBytes:  MaxHeaderBytes,
		userscontroller: UsersController,
	}
}

func (s *HttpServer) SetupRoutes() error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", s.userscontroller.GetUsers)
	mux.HandleFunc("POST /users", s.userscontroller.PostUsers)
	mux.HandleFunc("PUT /users", s.userscontroller.PutUsers)
	mux.HandleFunc("PATCH /users", s.userscontroller.PatchUsers)

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
