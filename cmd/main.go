package main

import (
	"encoding/json"
	"fmt"
	"os"

	repository "github.com/abhinash-kml/go-api-server/internal/repositories"
	service "github.com/abhinash-kml/go-api-server/internal/services"
)

func main() {
	// server := servers.NewHttpServer(":9000", time.Second*1, time.Second*1, time.Second*30, 2048)
	// server.SetupRoutes()
	// if err := server.Start(); err != nil {
	// 	fmt.Println("Error: ", err)
	// }

	repository := repository.NewInMemoryRepository()
	repository.Setup()
	userservice := service.NewLocalUserService(repository)

	users, err := userservice.GetUsers()
	if err != nil {
		fmt.Println(err.Error())
	}

	encoder := json.NewEncoder(os.Stdout)

	for _, value := range users {
		encoder.Encode(&value)
	}
}
