package servers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/abhinash-kml/go-api-server/config"
	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	m "github.com/abhinash-kml/go-api-server/internal/middlewares"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	"github.com/abhinash-kml/go-api-server/pkg/util"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Hook func() error

type CustomHttpServer struct {
	// Internal http server
	server     *http.Server
	mux        *http.ServeMux
	authConfig *config.AuthTokenConfig

	// Controllers
	userscontroller    controller.UsersController
	postscontroller    controller.PostsController
	commentscontroller controller.CommentsController

	// Logger
	logger zap.Logger

	// Before start hooks
	beforeStartHooks []Hook

	// After start hooks
	afterStartHooks []Hook

	// Before stop hooks
	beforeStopHooks []Hook

	// After stop hooks
	afterStopHooks []Hook
}

func NewHttpWithConfig(config *config.HttpConfig, authConfig *config.AuthTokenConfig, options ...FunctionalOption) *CustomHttpServer {
	internal := &http.Server{
		Addr:           fmt.Sprintf(":%s", config.Port),
		IdleTimeout:    time.Second * time.Duration(config.IdleTimeout),
		ReadTimeout:    time.Second * time.Duration(config.ReadTimeout),
		WriteTimeout:   time.Second * time.Duration(config.WriteTimeout),
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	wrapper := &CustomHttpServer{
		server:     internal,
		authConfig: authConfig,
	}

	for _, option := range options {
		option(wrapper)
	}

	if wrapper.mux != nil {
		wrapper.server.Handler = wrapper.mux
	} else {
		defaultmux := http.NewServeMux()
		wrapper.mux = defaultmux
		wrapper.server.Handler = wrapper.mux
	}

	return wrapper
}

func NewCustomCustomHttpServer(options ...FunctionalOption) *CustomHttpServer {
	internal := &http.Server{}
	wrapper := &CustomHttpServer{server: internal}

	for _, value := range options {
		value(wrapper)
	}

	// Mux provided externally, use that in internal server else provide default one
	if wrapper.mux != nil {
		wrapper.server.Handler = wrapper.mux
	} else {
		defaultmux := http.NewServeMux()     // Create default
		wrapper.mux = defaultmux             // Assign external pointer first
		wrapper.server.Handler = wrapper.mux // Assign internal
	}

	return wrapper
}

type FunctionalOption func(*CustomHttpServer)

func WithAddress(address string) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.server.Addr = address
	}
}

func WithReadTimeout(time time.Duration) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.server.ReadTimeout = time
	}
}

func WithWriteTimeout(time time.Duration) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.server.WriteTimeout = time
	}
}

func WithIdleTimeout(time time.Duration) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.server.IdleTimeout = time
	}
}

func WithMaxHeaderBytes(bytes int) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.server.MaxHeaderBytes = bytes
	}
}

func WithHandler(handler *http.ServeMux) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.mux = handler
	}
}

func WithTlsConfig(config *tls.Config) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.server.TLSConfig = config
	}
}

func WithLogger(logger zap.Logger) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.logger = logger
	}
}

func WithUsersController(controller controller.UsersController) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.userscontroller = controller
	}
}

func WithPostsController(controller controller.PostsController) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.postscontroller = controller
	}
}

func WithCommentsController(controller controller.CommentsController) FunctionalOption {
	return func(c *CustomHttpServer) {
		c.commentscontroller = controller
	}
}

type Myclaims struct {
	jwt.RegisteredClaims
	Myid string `json:"myid"`
	Meow string `json:"meow"`
}

func (s *CustomHttpServer) SetupDefaultRoutes() error {
	// Token routes
	s.mux.Handle("GET /login", m.CompileHandlers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenDuration, err := time.ParseDuration(s.authConfig.AccessToken.Expiration)
		if err != nil {
			log.Fatal("Failed to parse time duration for access token expiration")
		}
		refreshTokenDuration, err := time.ParseDuration(s.authConfig.RefreshToken.Expiration)
		if err != nil {
			log.Fatal("Failed to parse time duration for access token expiration")
		}

		s.logger.Debug("Access Token", zap.Float64("Duration", accessTokenDuration.Minutes()))
		s.logger.Debug("Refresh Token", zap.Float64("Duration", refreshTokenDuration.Minutes()))

		// Create access token
		accessToken, err := util.CreateJwtToken(
			s.authConfig.AccessToken.Secret,
			s.authConfig.AccessToken.Issuer,
			"123", // This subject should be filled with real entity identifier
			[]string{s.authConfig.AccessToken.Audience},
			"testid",
			accessTokenDuration,
		)
		if err != nil {
			s.logger.Info("Access Token creation error", zap.Error(err))
		}

		// Create refresh token
		refreshToken, err := util.CreateJwtToken(
			s.authConfig.RefreshToken.Secret,
			s.authConfig.RefreshToken.Issuer,
			"123", // This subject should be filled with real entity identifier
			[]string{s.authConfig.RefreshToken.Audience},
			"testid",
			refreshTokenDuration,
		)
		if err != nil {
			s.logger.Info("Refresh Token creation error", zap.Error(err))
		}

		response := model.AuthResponse{
			AccessToken:  accessToken,
			TokenType:    "Bearer",
			ExpiresIn:    10,
			RefreshToken: refreshToken,
			Scope:        "null",
		}

		s.logger.Debug("Auth Response", zap.String("Info", fmt.Sprintf("%+v", response)))
		json.NewEncoder(w).Encode(response)

	}), m.RateLimit, m.Logger))

	// Subjected to change
	s.mux.Handle("POST /access_token", m.CompileHandlers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody model.AccessTokenRequest
		json.NewDecoder(r.Body).Decode(&requestBody)

		// Verify jwt using util function
		_, claims, err := util.VerifyJwtToken(&s.authConfig.RefreshToken, requestBody.RefreshToken)
		if err != nil {
			s.logger.Info("Refresh token verfication failed", zap.Error(err))
		}

		// Check if Token ID i.e JTI is in ban list
		// using real logic like checking JTI from banlist store
		tokenId := claims.ID

		// This is placeholder logic
		fmt.Println("Got token id:", tokenId)

		// If banned then dont provide a new token
		// Else provide a new access token

	}), m.RateLimit, m.Logger))

	s.mux.Handle("GET /sse", m.CompileHandlers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Not supported", http.StatusInternalServerError)
			return
		}

		ticker := time.NewTicker(time.Second * 2)

		for {
			select {
			case <-r.Context().Done():
				{
					fmt.Println("Client Disconnected")
					return
				}
			case <-ticker.C:
				{
					time := time.Now().Format("2006-01-02 15:04:05")
					fmt.Fprintf(w, "data: Time: %s\n\n", time)
					flusher.Flush()
				}
			}
		}
	})))

	// Users routes
	s.mux.Handle("GET /users", m.CompileHandlers(http.HandlerFunc(s.userscontroller.GetUsers), m.JwtAuthorization, m.RateLimit, m.Logger)) // On test
	s.mux.Handle("GET /users/{id}", m.CompileHandlers(http.HandlerFunc(s.userscontroller.GetById), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("GET /users/{id}/posts", m.CompileHandlers(http.HandlerFunc(s.userscontroller.GetPostsOfUser), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("POST /users", m.CompileHandlers(http.HandlerFunc(s.userscontroller.PostUser), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("PUT /users", m.CompileHandlers(http.HandlerFunc(s.userscontroller.PutUser), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("PATCH /users", m.CompileHandlers(http.HandlerFunc(s.userscontroller.PatchUser), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("DELETE /users", m.CompileHandlers(http.HandlerFunc(s.userscontroller.DeleteUser), m.JwtAuthorization, m.RateLimit, m.Logger))

	// Post routes
	s.mux.Handle("GET /posts", m.CompileHandlers(http.HandlerFunc(s.postscontroller.GetPosts), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("GET /posts/{id}", m.CompileHandlers(http.HandlerFunc(s.postscontroller.GetById), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("GET /posts/{id}/comments", m.CompileHandlers(http.HandlerFunc(s.postscontroller.GetCommentsOfPost), m.JwtAuthorization, m.RateLimit, m.Logger)) // NEW
	s.mux.Handle("POST /posts", m.CompileHandlers(http.HandlerFunc(s.postscontroller.PostPost), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("PUT /posts", m.CompileHandlers(http.HandlerFunc(s.postscontroller.PutPost), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("PATCH /posts", m.CompileHandlers(http.HandlerFunc(s.postscontroller.PatchPost), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("DELETE /posts", m.CompileHandlers(http.HandlerFunc(s.postscontroller.DeletePost), m.JwtAuthorization, m.RateLimit, m.Logger))

	// Comments routes
	s.mux.Handle("GET /comments", m.CompileHandlers(http.HandlerFunc(s.commentscontroller.GetComments), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("GET /comments/{id}", m.CompileHandlers(http.HandlerFunc(s.commentscontroller.GetById), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("POST /comments", m.CompileHandlers(http.HandlerFunc(s.commentscontroller.PostComment), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("PUT /comments", m.CompileHandlers(http.HandlerFunc(s.commentscontroller.PutComment), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("PATCH /comments", m.CompileHandlers(http.HandlerFunc(s.commentscontroller.PatchComment), m.JwtAuthorization, m.RateLimit, m.Logger))
	s.mux.Handle("DELETE /comments", m.CompileHandlers(http.HandlerFunc(s.commentscontroller.DeleteComment), m.JwtAuthorization, m.RateLimit, m.Logger))

	return nil
}

func (s *CustomHttpServer) AddRoute(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *CustomHttpServer) AddRoutes() {

}

func (s *CustomHttpServer) AddBeforeStartHook(hook Hook) {
	s.beforeStartHooks = append(s.beforeStartHooks, hook)
}

func (s *CustomHttpServer) AddAfterStartHook(hook Hook) {
	s.afterStartHooks = append(s.afterStartHooks, hook)
}

func (s *CustomHttpServer) AddBeforeStopHook(hook Hook) {
	s.beforeStopHooks = append(s.beforeStopHooks, hook)
}

func (s *CustomHttpServer) AddAfterStopHook(hook Hook) {
	s.afterStopHooks = append(s.afterStopHooks, hook)
}

func (s *CustomHttpServer) Start() error {
	// Execute before start hooks
	for _, hooks := range s.beforeStartHooks {
		if err := hooks(); err != nil {
			s.logger.Error("Failed to start server", zap.Error(err))
			return err
		}
	}

	fmt.Println("Starting HTTP server on ", s.server.Addr)

	// Start the actual server on another goroutine and listen for error on buffered channel
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.server.ListenAndServe()
	}()

	// Lets wait for an immediate failure within 2 sec and then execute after start hooks
	select {
	case <-time.After(time.Second * 2):
		{
			s.logger.Info("Server started successfully.", zap.String("Address", s.server.Addr))

			// Execute after start hooks in their own goroutine so we start monitoring for error immediately
			go func() {
				for _, hooks := range s.afterStartHooks {
					if err := hooks(); err != nil {
						s.logger.Error("After start hook error", zap.Error(err))
					}
				}
			}()

			// Monitor errChan in background goroutine as now Start() is returning
			// and its a blocking work
			go func() {
				err := <-errChan
				if err != nil && err != http.ErrServerClosed {
					s.logger.Error("Server crashed after successful start", zap.Error(err))
				}
			}()

			return nil // Success
		}
	case err := <-errChan: // Server failed within 2 secs
		return err
	}
}

func (s *CustomHttpServer) Stop() error {
	// Execute before stop hooks
	for _, hooks := range s.beforeStopHooks {
		if err := hooks(); err != nil {
			s.logger.Error("Failed to stop server", zap.Error(err))
			return err
		}
	}

	fmt.Println("Shutting down http server...")

	// Give it a grace period of 2 secs to terminate all connections and free up resources
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	s.server.Shutdown(ctx)

	// Execute after stop hooks
	for _, hooks := range s.afterStopHooks {
		if err := hooks(); err != nil {
			s.logger.Error("After stop hook error", zap.Error(err))
			return err
		}
	}

	return nil
	// os.Exit(0)
}
