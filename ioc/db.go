package ioc

import (
	"webook/config"
	"webook/internal/repository/dao"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.AppConfig.DB.DSN))
	if err != nil {
		panic(err)
	}

	if err = dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}
