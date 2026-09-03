package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	return dao.db.WithContext(ctx).Create(&u).Error
}

// User 直接对应数据库表结构
// 有些人叫做 entity， 有些人叫做 model, 有些人叫做 PO(Persistent Objext)
type User struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	// 唯一索引
	Email    string `gorm:"unique"`
	Password string

	// 增加字段

	// 创建时间， 毫秒数
	// 不用Time.time： 时区问题处理很麻烦
	Ctime int64
	// 更新时间， 毫秒数
	Utime int64
}

