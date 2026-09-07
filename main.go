package main

import (
	"net/http"
	"strings"
	"time"

	"webbook/config"
	"webbook/internal/repository"
	"webbook/internal/repository/dao"
	"webbook/internal/service"
	"webbook/internal/web"
	"webbook/internal/web/middleware"
	"webbook/pkg/ginx/middlewares/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	sessionredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	u := initUser(db)
	u.RegisterRoutes(server)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, welcone")
	})
	server.Run(":8080")
}

// func initRedis() *redis.Client {
// 	return redis.NewClient(&redis.Options{
// 		Addr:                  "localhost:6379",
// 		PoolSize:              100,
// 		MinIdleConns:          16,
// 		ConnMaxIdleTime:       5 * time.Minute,
// 		DialTimeout:           time.Second,
// 		ReadTimeout:           time.Second,
// 		WriteTimeout:          time.Second,
// 		PoolTimeout:           time.Second,
// 		ContextTimeoutEnabled: true,
// 		// 限流脚本会写入记录，关闭自动重试以避免重复执行。
// 		MaxRetries: -1,
// 	})
// }

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(func(ctx *gin.Context) {
		println("这是第一个middleware")
	})

	server.Use(func(ctx *gin.Context) {
		println("这是第二个middleware")
	})

	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://localhost:3000"},
		// AllowMethods:     []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 不加这个前端拿不到（我给你的你才能要）
		ExposeHeaders: []string{"x-jwt-token"},
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.AppConfig.Redis.Addr,
	})

	// 每个 IP 在任意一分钟内最多通过 100 次请求。
	server.Use(ratelimit.NewBuilder(redisClient, time.Minute, 100).Build())

	// 步骤1
	// session的数据存哪里
	// 一个基于cookie的store实现
	// store := cookie.NewStore([]byte("secret"))

	// 两个Key最好是32bit或者64bit
	// store := memstore.NewStore([]byte("authenticationKey"), []byte("encryptionKey"))

	store, err := sessionredis.NewStore(16,
		"tcp", config.AppConfig.Redis.Addr, "root", "",
		[]byte("51c78d409996e61725278ee9a4dee314eba8da064211e31aa4b915412f438ae8"),
		[]byte("b8ab129bcbf47d6a7dea78eda2820e37"))
	if err != nil {
		panic(err)
	}
	// cookie的对应的位置
	server.Use(sessions.Sessions("mysession", store))
	// 步骤3 Builder模式的优越性
	// server.Use(middleware.NewLoginMiddlewareBuilder().
	// 	IgnorePaths("/users/signup").
	// 	IgnorePaths("/users/login").Buid())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		Buid())

	// // v1
	// middleware.IgnorePaths = []string{"sss"}
	// server.Use(middleware.CheckLogin())

	// // 不能忽略sss这条路径
	// server1 := gin.Default()
	// server1.Use(middleware.CheckLogin())
	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	dao := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.AppConfig.DB.DSN))
	if err != nil {
		// 只会在初始化过程中panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错， 应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
