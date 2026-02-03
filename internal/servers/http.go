package servers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	"github.com/golang-jwt/jwt"
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

type Myclaims struct {
	jwt.StandardClaims
	Myid string `json:"myid"`
	Meow string `json:"meow"`
}

func (s *HttpServer) SetupRoutes() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
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
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				Issuer:    "Neo",
				ExpiresAt: time.Now().Add(time.Hour * 2).Unix(),
				Subject:   "nice subject",
				Id:        "nkaheui",
				Audience:  "Hello Audience",
			},
		}

		// Send JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, mclaims)
		signedToken, _ := token.SignedString([]byte("my-secret-key"))
		w.Write([]byte(fmt.Sprintf("Jwt Token: %s", signedToken)))
	})

	mux.HandleFunc("GET /verify", func(w http.ResponseWriter, r *http.Request) {
		authheader := r.Header.Get("Authorization")
		separated := strings.Split(authheader, " ")

		if len(separated) < 2 {
			http.Error(w, "Bad Header", http.StatusBadRequest)
		}

		if separated[0] != "Bearer" {
			http.Error(w, "Bad Token Scheme", http.StatusUnauthorized)
		}

		token, _ := jwt.ParseWithClaims(separated[1], &Myclaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte("my-secret-key"), nil
		})

		if claims, ok := token.Claims.(Myclaims); ok && token.Valid {
			if claims.Meow != "llll" || claims.Myid != "mmm" || claims.Issuer != "Neo" {
				http.Error(w, "Invalid claims", 401)
				return
			}
		}

		w.Write([]byte("Success"))
	})

	mux.HandleFunc("GET /offset", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("GET /cursor", func(w http.ResponseWriter, r *http.Request) {
		nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		after := r.URL.Query().Get("after")
		before := r.URL.Query().Get("before")

		fmt.Println("Page:", page)
		fmt.Println("After:", after)
		fmt.Println("Before:", before)
		fmt.Println(nums)
	})

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
