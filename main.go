package main

import (
	"net/http"
	"strings"
	"time"

	"webbook/config"
	"webbook/internal/repository"
	"webbook/internal/repository/cache"
	"webbook/internal/repository/dao"
	"webbook/internal/service"
	"webbook/internal/service/sms"
	"webbook/internal/service/sms/tencent"
	"webbook/internal/web"
	"webbook/internal/web/middleware"
	"webbook/pkg/ginx/middlewares/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	sessionredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	smsv20210111 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	// redisClient 只建一次，限流、验证码缓存、用户缓存共用一个连接池
	redisClient := initRedis()
	server := initWebServer(redisClient)
	u := initUser(db, redisClient)
	u.RegisterRoutes(server)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, welcone")
	})
	server.Run(":8080")
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: config.AppConfig.Redis.Addr,
	})
}

func initSMSService() sms.Service {
	// 腾讯云的密钥和地域
	credential := common.NewCredential(
		config.AppConfig.SMS.SecretId,
		config.AppConfig.SMS.SecretKey,
	)
	client, err := smsv20210111.NewClient(credential,
		"ap-guangzhou", profile.NewClientProfile())
	if err != nil {
		// 初始化过程中出错，就不要启动了
		panic(err)
	}
	return tencent.NewService(client,
		config.AppConfig.SMS.AppId, config.AppConfig.SMS.SignName)
}

func initWebServer(redisClient *redis.Client) *gin.Engine {
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

	// redisClient 由 main 传进来，不要在这里初始化
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
	// 发验证码的时候还没登录，不能拦
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		Buid())

	// // v1
	// middleware.IgnorePaths = []string{"sss"}
	// server.Use(middleware.CheckLogin())

	// // 不能忽略sss这条路径
	// server1 := gin.Default()
	// server1.Use(middleware.CheckLogin())
	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(userCache, userDAO)
	svc := service.NewUserService(repo)

	// 验证码：cache -> repository -> service，短信服务注入进去
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	codeSvc := service.NewCodeService(codeRepo, initSMSService())

	u := web.NewUserHandler(svc, codeSvc)
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
