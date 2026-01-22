package main

import (
	"fmt"
	"time"

	"github.com/abhinash-kml/go-api-server/internal/servers"
)

func main() {
	fmt.Println("Hello World")
	server := servers.NewHttpServer(":8080", time.Second*1, time.Second*1, time.Second*30, 2048)
	server.SetupRoutes()
	server.Start()
}
