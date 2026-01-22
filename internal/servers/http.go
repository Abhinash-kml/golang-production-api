package servers

import (
	"crypto/tls"
	"fmt"
	"net/http"
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
	logger         zap.Logger
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

func (s *HttpServer) SetupRoutes() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", controller.GetUsers)
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
