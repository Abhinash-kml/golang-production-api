package servers

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Hook func() error

type CustomHttpServer struct {
	// Internal http server
	server *http.Server
	mux    *http.ServeMux

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

func (s *CustomHttpServer) SetupRoutes() error {
	s.mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Custom-Header")
		w.Header().Set("Access-Control-Max-Age", "86400")

		mclaims := Myclaims{
			Myid: "mmm",
			Meow: "llll",
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "Neo",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 2)),
				Subject:   "nice subject",
				ID:        "nkaheui",
				Audience:  []string{"Hello Audience"},
			},
		}

		// Send JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, mclaims)
		signedToken, _ := token.SignedString([]byte("my-secret-key"))
		w.Write([]byte(fmt.Sprintf("Jwt Token: %s", signedToken)))
	})

	s.mux.HandleFunc("GET /verify", func(w http.ResponseWriter, r *http.Request) {
		authheader := r.Header.Get("Authorization")
		separated := strings.Split(authheader, " ")

		if len(separated) < 2 || separated[0] != "Bearer" {
			http.Error(w, "Bad Header", http.StatusUnauthorized)
		}

		keyfunc := func(token *jwt.Token) (interface{}, error) {
			return []byte("my-secret-key"), nil
		}

		token, _ := jwt.ParseWithClaims(separated[1], &Myclaims{}, keyfunc)

		if claims, ok := token.Claims.(Myclaims); ok && token.Valid {
			if claims.Meow != "llll" || claims.Myid != "mmm" || claims.Issuer != "Neo" {
				http.Error(w, "Invalid claims", 401)
				return
			}
		}

		w.Write([]byte("Success"))
	})

	s.mux.HandleFunc("GET /offset", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		offset, _ := strconv.Atoi(queryParams.Get("offset"))
		limit, _ := strconv.Atoi(queryParams.Get("limit"))
		sort := queryParams.Get("sort")

		nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

		if offset < 0 || offset >= 20 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
		}

		if limit <= 0 || limit > 10 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
		}

		payload := nums[offset : offset+limit]
		switch sort {
		case "asc":
			json.NewEncoder(w).Encode(payload)
		case "desc":
			{
				slices.Reverse(payload)
				json.NewEncoder(w).Encode(payload)
			}
		}
	})

	s.mux.HandleFunc("GET /cursor", func(w http.ResponseWriter, r *http.Request) {
		nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		after := r.URL.Query().Get("after")
		before := r.URL.Query().Get("before")

		var (
			encodedAfter  string
			encodedBefore string
			// decodedAfter  []byte
			// decodedBefore []byte
		)

		// base64.URLEncoding.Decode(decodedAfter, []byte(after))
		// base64.URLEncoding.Decode(decodedBefore, []byte(before))
		encodedAfter = base64.URLEncoding.EncodeToString([]byte("20hgwfhbkewfnk,ejwfnewf"))
		encodedBefore = base64.URLEncoding.EncodeToString([]byte("10,mfbnewkjufrhewkfnwfjbqedq"))

		fmt.Println("Encoded after:", string(encodedAfter))
		fmt.Println("Encoded before:", string(encodedBefore))

		fmt.Println("Limit:", limit)
		fmt.Println("After:", after)
		fmt.Println("Before:", before)
		fmt.Println(nums)
	})

	// Users routes
	s.mux.HandleFunc("GET /users", s.userscontroller.GetUsers)
	s.mux.HandleFunc("POST /users", s.userscontroller.PostUsers)
	s.mux.HandleFunc("PUT /users", s.userscontroller.PutUsers)
	s.mux.HandleFunc("PATCH /users", s.userscontroller.PatchUsers)

	// Post routes
	s.mux.HandleFunc("GET /posts", s.postscontroller.GetPosts)
	s.mux.HandleFunc("POST /posts", s.postscontroller.PostPosts)
	s.mux.HandleFunc("PUT /posts", s.postscontroller.PutPosts)
	s.mux.HandleFunc("PATCH /posts", s.postscontroller.PatchPosts)

	// Comments routes
	s.mux.HandleFunc("GET /comments", s.commentscontroller.GetComments)
	s.mux.HandleFunc("POST /comments", s.commentscontroller.PostComments)
	s.mux.HandleFunc("PUT /comments", s.commentscontroller.PutComments)
	s.mux.HandleFunc("PATCH /comments", s.commentscontroller.PatchComments)

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
