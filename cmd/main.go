package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/abhinash-kml/go-api-server/config"
	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	model "github.com/abhinash-kml/go-api-server/internal/models"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/abhinash-kml/go-api-server/internal/servers"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Signal handling
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)

	// Config
	config := config.Initialize()

	// Log file
	logFile, err := os.OpenFile("./logs/temp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file")
	}
	defer logFile.Close()

	// Logger - stdout + file
	fileSyncer := zapcore.AddSync(logFile)
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	loglevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileSyncer, loglevel)
	stdoutCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), stdoutSyncer, loglevel)
	combinedCore := zapcore.NewTee(fileCore, stdoutCore)
	logger := zap.New(combinedCore, zap.AddCaller())
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password
		DB:       0,  // Default db
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			fmt.Println("Connected to redis")
			return nil
		},
	})

	data := map[string]string{
		"id":      "100",
		"name":    "neo",
		"city":    "Kolkata",
		"state":   "West Bengal",
		"country": "India",
	}
	rdb.HSet(context.Background(), "user", data)
	test := model.User{}
	if err := rdb.HGetAll(context.Background(), "user").Scan(&test); err != nil {
		fmt.Println("Error:", err.Error())
	}
	fmt.Printf("%+v", test)
	// rdb.HDel(context.Background(), "user")

	// Repository
	userrepository := repository.NewInMemoryUsersRepository()
	userrepository.Setup()
	postsrepository := repository.NewInMemoryPostsRepository()
	postsrepository.Setup()
	commentrepository := repository.NewInMemoryCommentsRepository()
	commentrepository.Setup()

	// Service
	userservice := service.NewLocalUserService(userrepository, rdb)
	postsservice := service.NewLocalPostsService(postsrepository, rdb)
	commentservice := service.NewLocalCommentService(commentrepository, rdb)

	// Controllers
	usercontroller := controller.NewUsersController(userservice, postsservice, commentservice, logger)
	postscontroller := controller.NewPostsController(userservice, postsservice, commentservice, logger)
	commentscontroller := controller.NewCommentsController(userservice, postsservice, commentservice, logger)

	// server := servers.NewCustomCustomHttpServer(
	// 	servers.WithAddress(":9000"),
	// 	servers.WithIdleTimeout(time.Second*15),
	// 	servers.WithReadTimeout(time.Second*15),
	// 	servers.WithWriteTimeout(time.Second*5),
	// 	servers.WithMaxHeaderBytes(1500),
	// 	servers.WithLogger(*logger),
	// 	servers.WithUsersController(*usercontroller),
	// 	servers.WithPostsController(*postscontroller),
	// 	servers.WithCommentsController(*commentscontroller))

	server := servers.NewHttpWithConfig(&config.Server.Http,
		&config.Auth,
		servers.WithLogger(*logger),
		servers.WithUsersController(*usercontroller),
		servers.WithPostsController(*postscontroller),
		servers.WithCommentsController(*commentscontroller))

	server.SetupDefaultRoutes()
	server.AddRoute("GET /custom", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Registering custom routes successfull\nIts working yeaaahhh"))
	})

	server.AddBeforeStartHook(func() error {
		fmt.Println("Before start hook...")
		return nil
	})

	server.AddAfterStartHook(func() error {
		fmt.Println("After start hook...")
		return nil
	})

	server.AddBeforeStopHook(func() error {
		fmt.Println("Before stop hook...")
		return nil
	})

	server.AddAfterStopHook(func() error {
		fmt.Println("After stop hook...")
		return nil
	})

	if err := server.Start(); err != nil {
		fmt.Println("Error: ", err.Error())
	}

	// limiter := ratelimiter.FixedWindowLimiter{
	// 	WindowDuration: time.Second * 10,
	// 	LimitPerWindow: 5,
	// 	Table:          make(map[string]*ratelimiter.ClientInfo),
	// }

	// go func() {
	// 	count := 0
	// 	for {
	// 		if limiter.Allow("aaa") {
	// 			count++
	// 			fmt.Println("Allowed:", count)
	// 		} else {
	// 			continue
	// 		}
	// 	}
	// }()

	fmt.Println("Listening for termination syscall...")
	fmt.Println("Got:", <-stopSig)
	server.Stop()
}
