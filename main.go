package main

import (
	"webbook/internal/web"
)

func main() {
	server := web.RegiterRoutes()

	u := web.NewUserHandler()
	u.RegisterRoutes(server)

	server.Run(":8080")
}
