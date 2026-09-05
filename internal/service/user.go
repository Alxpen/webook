package service

import (
	"context"
	"errors"

	"webbook/internal/domain"
	"webbook/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
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

func (svc *UserService) UpdateNonSensitiveInfo(
	ctx context.Context,
	u domain.User,
) error {
	return svc.repo.UpdateNonSensitiveInfo(ctx, u)
}

func (svc *UserService) Profile(
	ctx context.Context,
	id int64,
) (domain.User, error) {
	return svc.repo.FindByID(ctx, id)
}
