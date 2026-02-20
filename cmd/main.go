package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/abhinash-kml/go-api-server/config"
	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/abhinash-kml/go-api-server/internal/servers"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.uber.org/zap"
)

func main() {
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)

	config := config.Initialize()

	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		panic("Unable to initialize logger - Zap")
	}

	userrepository := repository.NewInMemoryUsersRepository()
	userrepository.Setup()
	postsrepository := repository.NewInMemoryPostsRepository()
	postsrepository.Setup()
	commentrepository := repository.NewInMemoryCommentsRepository()
	commentrepository.Setup()

	userservice := service.NewLocalUserService(userrepository)
	postsservice := service.NewLocalPostsService(postsrepository)
	commentservice := service.NewLocalCommentService(commentrepository)

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
