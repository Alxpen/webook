package service

import (
	"context"
	"webbook/internal/domain"
	"webbook/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// domain.User用指针的话需要判空
func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 考虑加密放在哪里的问题

	// 然后就是存起来
	return svc.repo.Create(ctx, u)
}
