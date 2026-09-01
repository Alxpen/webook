package web

import (
	"github.com/gin-gonic/gin"
)

func RegiterRoutes() *gin.Engine {
	server := gin.Default()
	// RegisterUserRoutes(server)
	return server
}

// func RegisterUserRoutes(server *gin.Engine) {	
// }
