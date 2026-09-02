package main

import (
	"strings"
	"time"
	"webbook/internal/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := web.RegiterRoutes()

	server.Use(func(ctx *gin.Context) {
		println("这是第一个middleware")
	})

	server.Use(func(ctx *gin.Context) {
		println("这是第二个middleware")
	})

	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://localhost:3000"},
		// AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		// 是否允许你带认证信息
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))
	u := web.NewUserHandler()
	u.RegisterRoutes(server)

	server.Run(":8080")
}
