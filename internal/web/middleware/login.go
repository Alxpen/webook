package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

// 方法1
func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// 可以在登录校验的时候刷新session
func (l *LoginMiddlewareBuilder) Buid() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		// if ctx.Request.URL.Path == "/users/login" || ctx.Request.URL.Path == "/users/signup" {
		// 	return
		// }
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		sess.Options(sessions.Options{
			MaxAge: 60 * 30,
		})
		now := time.Now().UnixMilli()
		// 说明没有刷新过,刚登录
		if updateTime == nil {
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "刷新登录状态失败",
				})
				return
			}
			return
		}
		// updateTime 是有的
		updateTimeVal, _ := updateTime.(int64)
		if now-updateTimeVal > 60*1000 {
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "刷新登录状态失败",
				})
				return
			}
		}
	}
}

// // 方法2
// var IgnorePaths []string

// func CheckLogin() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		for _, path := range IgnorePaths {
// 			if ctx.Request.URL.Path == path {
// 				return
// 			}
// 		}
// 		sess := sessions.Default(ctx)
// 		id := sess.Get("userId")
// 		if id == nil {
// 			ctx.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}
// 	}
// }

// // 方法3
// func CheckLoginV1(paths []string) gin.HandlerFunc {
// 	if len(paths) == 0 {
// 		paths = []string{}
// 	}
// 	return func(ctx *gin.Context) {
// 		for _, path := range paths {
// 			if ctx.Request.URL.Path == path {
// 				return
// 			}
// 		}
// 		sess := sessions.Default(ctx)
// 		id := sess.Get("userId")
// 		if id == nil {
// 			ctx.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}
// 	}
// }
