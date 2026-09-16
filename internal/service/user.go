package service

import (
	"context"
	"errors"

	"webbook/internal/domain"
	"webbook/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("邮箱或密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

// domain.User用指针的话需要判空
func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 考虑加密放在哪里的问题
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	// 然后就是存起来
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindByID(ctx, id)
	return u, err
}

// FindOrCreate 如果手机号不存在，那么会初始化一个用户
func (svc *UserService) FindOrCreate(ctx context.Context,
	phone string,
) (domain.User, error) {
	// 这是一种优化写法
	// 大部分人会命中这个分支
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 要执行注册
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	// 注册有问题，但是又不是用户手机号码冲突，说明是系统错误
	if err != nil && err != repository.ErrUserDuplicate {
		return domain.User{}, err
	}
	// 主从模式下，这里要从主库中读取，暂时我们不需要考虑
	return svc.repo.FindByPhone(ctx, phone)
}
