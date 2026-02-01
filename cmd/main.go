package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	controller "github.com/abhinash-kml/go-api-server/internal/controllers"
	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	"github.com/abhinash-kml/go-api-server/internal/servers"
	service "github.com/abhinash-kml/go-api-server/internal/services"
	"go.uber.org/zap"
)

func main() {
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewProduction()
	if err != nil {
		panic("Unable to initialize logger - Zap")
	}

	userrepository := repository.NewInMemoryRepository()
	userrepository.Setup()
	userservice := service.NewLocalUserService(userrepository)
	usercontroller := controller.NewUsersController(userservice, logger)

	commentrepository := repository.NewInMemoryCommentsRepository()
	commentrepository.Setup()
	commentservice := service.NewLocalCommentService(commentrepository)
	commentscontroller := controller.NewCommentsController(commentservice, logger)

	server := servers.NewHttpServer(
		":9000",
		time.Second*1,
		time.Second*1,
		time.Second*30,
		2048,
		*usercontroller,
		*commentscontroller)

	server.SetupRoutes()
	go func() {
		if err := server.Start(); err != nil {
			fmt.Println("Error: ", err.Error())
		}
	}()

	<-stopSig
	server.Stop()
}
