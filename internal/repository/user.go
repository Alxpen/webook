package repository

import (
	"context"
	"time"

	"webbook/internal/domain"
	"webbook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

// var ErrUserDuplicateEmailV1 = fmt.Errorf("%w 邮箱冲突", dao.ErrUserDuplicateEmail)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
	// 在这里操作缓存
}

func (r *UserRepository) UpdateNonSensitiveInfo(
	ctx context.Context,
	u domain.User,
) error {
	return r.dao.UpdateUnsensetiveInfo(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
	})
}

func (r *UserRepository) FindByID(
	ctx context.Context,
	id int64,
) (domain.User, error) {
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	var birthday time.Time
	if u.Birthday > 0 {
		birthday = time.UnixMilli(u.Birthday).UTC()
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Nickname: u.Nickname,
		Birthday: birthday,
		AboutMe:  u.AboutMe,
	}, nil
}
