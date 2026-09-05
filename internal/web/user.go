package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"webbook/internal/domain"
	"webbook/internal/repository"
	"webbook/internal/service"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UserHandler 准备在上面定义跟用户相关的路由
type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	// 预编译正则表达式
	const (
		emailRegexPattern    = `^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,72}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", u.Profile)
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind方法会根据Content-Type 来解析你的数据到req中
	// 解析错误就直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不对")
		return
	}
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次输入密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于八位，包含数字、特殊字符")
		return
	}
	// 调用UserService
	err = u.svc.SignUp(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// errors.Is()
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
	fmt.Printf("%v", req)
	// 数据库操作desuwa
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 登录成功之后取出session
	// 设置session
	sess := sessions.Default(ctx)
	// 随便设置放在session里面的值
	sess.Set("userId", user.Id)
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req EditReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "请求数据格式不正确",
		})
		return
	}

	nickname := strings.TrimSpace(req.Nickname)
	nicknameLength := utf8.RuneCountInString(nickname)
	if nicknameLength < 2 || nicknameLength > 20 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "昵称长度必须为2到20个字符",
		})
		return
	}

	const dateLayout = "2006-01-02"
	birthday, err := time.Parse(dateLayout, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "生日格式必须为YYYY-MM-DD，例如1992-01-01",
		})
		return
	}

	earliestBirthday := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if birthday.Before(earliestBirthday) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "生日不能早于1900-01-01",
		})
		return
	}

	if birthday.After(time.Now().UTC()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "生日不能晚于今天",
		})
		return
	}

	aboutMe := strings.TrimSpace(req.AboutMe)
	if utf8.RuneCountInString(aboutMe) > 1024 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "个人简介不能超过1024个字符",
		})
		return
	}

	id, ok := sessions.Default(ctx).Get("userId").(int64)
	if !ok || id <= 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "登录状态无效，请重新登录",
		})
		return
	}

	err = u.svc.UpdateNonSensitiveInfo(
		ctx.Request.Context(),
		domain.User{
			Id:       id,
			Nickname: nickname,
			Birthday: birthday,
			AboutMe:  aboutMe,
		},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "保存个人信息失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "个人信息修改成功",
	})
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	id, ok := sessions.Default(ctx).Get("userId").(int64)
	if !ok || id <= 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "登录状态无效，请重新登录",
		})
		return
	}

	user, err := u.svc.Profile(ctx.Request.Context(), id)
	if errors.Is(err, repository.ErrUserNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "用户不存在",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取个人信息失败",
		})
		return
	}

	birthday := ""
	if !user.Birthday.IsZero() {
		birthday = user.Birthday.UTC().Format("2006-01-02")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":       user.Id,
		"email":    user.Email,
		"nickname": user.Nickname,
		"birthday": birthday,
		"aboutMe":  user.AboutMe,
	})
}
