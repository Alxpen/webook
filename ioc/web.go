package ioc

import (
	"strings"
	"time"

	"webook/config"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middlewares/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	sessionredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitGin(mdls []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient *redis.Client) []gin.HandlerFunc {
	store, err := sessionredis.NewStore(
		16,
		"tcp", config.AppConfig.Redis.Addr, "root", "",
		[]byte("51c78d409996e61725278ee9a4dee314eba8da064211e31aa4b915412f438ae8"),
		[]byte("b8ab129bcbf47d6a7dea78eda2820e37"),
	)
	if err != nil {
		panic(err)
	}

	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"x-jwt-token"},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return strings.HasPrefix(origin, "http://localhost") ||
					strings.Contains(origin, "yourcompany.com")
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(redisClient, time.Minute, 100).Build(),
		sessions.Sessions("mysession", store),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			Buid(),
	}
}
