package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	return u, err
}

func (dao *UserDAO) UpdateUnsensetiveInfo(ctx context.Context, u User) error {
	return dao.db.WithContext(ctx).Model(&User{}).
		Where("id=?", u.Id).
		Updates(map[string]any{
			"nickname": u.Nickname,
			"birthday": u.Birthday,
			"about_me": u.AboutMe,
			"utime":    time.Now().Unix(),
		}).Error
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	// 与底层MySQL底层强耦合，无可避免
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

// User 直接对应数据库表结构
// 有些人叫做 entity， 有些人叫做 model, 有些人叫做 PO(Persistent Objext)
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 唯一索引
	Email    string `gorm:"unique"`
	Password string

	// 增加字段
	Nickname string `gorm:"type:varchar(64)"`
	Birthday int64
	AboutMe  string `gorm:"type:varchar(1024)"`

	// 创建时间， 毫秒数
	// 不用Time.time： 时区问题处理很麻烦
	Ctime int64
	// 更新时间， 毫秒数
	Utime int64
}
