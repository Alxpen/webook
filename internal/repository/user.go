package repository

import (
	"context"
	"database/sql"

	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrUserNotFound  = dao.ErrUserNotFound
	ErrUserDuplicate = dao.ErrUserDuplicate
)

// var ErrUserDuplicateEmailV1 = fmt.Errorf("%w 邮箱冲突", dao.ErrUserDuplicateEmail)

type UserRepository struct {
	cache *cache.UserCache
	dao   *dao.UserDAO
}

func NewUserRepository(cache *cache.UserCache, dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		cache: cache,
		dao:   dao,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
	})
	// 在这里操作缓存
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// 直接返回
		return u, err
	}

	// if err == repository.ErrUserNotFound {
	// 	// 去数据库找
	// }

	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       ue.Id,
		Email:    ue.Email.String,
		Phone:    ue.Phone.String,
		Password: ue.Password,
	}

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 打个日志就好
		}
	}()

	return u, err
	// 这种情况怎么办
	// error == io.EOF
	// 面对两种情况，是否选择加载数据库？
	// 1、redis崩溃
	// 2、redis偶发性错误

	// 选择加载--做好兜底，万一redis崩了，要保护好数据库
	// 我数据库限流
	// 搞个布隆过滤器

	// 选择不加载--redis偶发性错误，降低体验
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
	}, nil
}
