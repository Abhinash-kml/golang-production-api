package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/abhinash-kml/go-api-server/config"
	"github.com/abhinash-kml/go-api-server/internal/connections"
	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	"github.com/abhinash-kml/go-api-server/internal/observability"
	"github.com/abhinash-kml/go-api-server/internal/realtime"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/abhinash-kml/go-api-server/internal/servers"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Signal handling
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)

	// Config
	config := config.Get()

	// Log file
	logFile, err := os.OpenFile("./logs/temp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file")
	}
	defer logFile.Close()

	// Logger - stdout + file
	fileSyncer := zapcore.AddSync(logFile)
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	loglevel := zap.NewAtomicLevelAt(zap.DebugLevel) // TODO: Parse and set log level from config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileSyncer, loglevel)
	stdoutCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), stdoutSyncer, loglevel)
	combinedCore := zapcore.NewTee(fileCore, stdoutCore)
	logger := zap.New(combinedCore, zap.AddCaller())
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Set up OpenTelemetry.
	otelShutdown, err := observability.SetupOTelSDK(context.Background())
	if err != nil {
		logger.Fatal("Failed setting up observability", zap.Error(err))
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Create tracers per domain
	usersTracer := otel.Tracer("users")
	postsTracer := otel.Tracer("posts")
	commentsTracer := otel.Tracer("comments")

	// Connections
	postgresdsn := "postgresql://postgres:Abx305@localhost:5432/goapp?sslmode=disable"
	postgresConnection := connections.NewPostgresConnection(postgresdsn)
	redisConnection := connections.NewRedisConnection(&redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: "",
	})

	// Migrations
	migrateFlag := flag.String("migrate", "none", "Usage: up, down, none (default)")
	flag.Parse()
	fmt.Println(migrateFlag) // Dummy
	m, err := migrate.New(
		"file://db/migrations",
		postgresdsn,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Perform schema migration as per action
	switch *migrateFlag {
	case "up":
		if err := m.Up(); err != nil {
			logger.Fatal("Migrate up failed", zap.Error(err))
		}
	case "down":
		if err := m.Down(); err != nil {
			logger.Fatal("Migrate down failed", zap.Error(err))
		}
	case "none": // Default case do nothing
	}

	// Repository
	userrepository := repository.NewPostgresUserRepository(postgresConnection, usersTracer)
	userrepository.Setup()
	postsrepository := repository.NewPostgresPostRepository(postgresConnection, postsTracer)
	postsrepository.Setup()
	commentrepository := repository.NewPostgresCommentRepository(postgresConnection, commentsTracer)
	commentrepository.Setup()

	// Service
	userservice := service.NewLocalUserService(userrepository, redisConnection, usersTracer)
	postsservice := service.NewLocalPostsService(postsrepository, redisConnection, postsTracer)
	commentservice := service.NewLocalCommentService(commentrepository, redisConnection, commentsTracer)

	// Controllers
	usercontroller := controller.NewUsersController(userservice, postsservice, commentservice, logger, usersTracer)
	postscontroller := controller.NewPostsController(userservice, postsservice, commentservice, logger, postsTracer)
	commentscontroller := controller.NewCommentsController(userservice, postsservice, commentservice, logger, commentsTracer)

	// Session store
	sessionstore := realtime.NewInMemorySessionStore()

	// Pub sub
	redisPubSub := realtime.NewRedisPubSub(redisConnection)

	// Hub
	hub := realtime.NewHub(sessionstore, &redisPubSub, realtime.PubSubTypeMemory)

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

	// Main http server
	server := servers.NewHttpWithConfig(&config.Server.Http,
		&config.Auth,
		servers.WithLogger(*logger),
		servers.WithUsersController(*usercontroller),
		servers.WithPostsController(*postscontroller),
		servers.WithCommentsController(*commentscontroller),
		servers.WithHub(hub))

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

	// stream := sse.NewClient("http://localhost:9000/sse")
	// time.AfterFunc(time.Second*3, func() {
	// 	go func() {
	// 		stream.Subscribe("message", func(msg *sse.Event) {
	// 			fmt.Println(string(msg.Data))
	// 		})
	// 	}()
	// })

	fmt.Println("Listening for termination syscall...")
	fmt.Println("Got:", <-stopSig)
	server.Stop()
}
