package main

import (
	"fmt"
	"time"

	"github.com/abhinash-kml/go-api-server/internal/servers"
)

func main() {
	server := servers.NewHttpServer(":9000", time.Second*1, time.Second*1, time.Second*30, 2048)
	server.SetupRoutes()
	if err := server.Start(); err != nil {
		fmt.Println("Error: ", err)
	}
}
