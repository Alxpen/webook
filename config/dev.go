//go:build !k8s

package config

var AppConfig = Config{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:13316)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
	// 腾讯云短信，去控制台拿自己账号的值填进来
	SMS: SMSConfig{
		AppId:     "",
		SignName:  "",
		SecretId:  "",
		SecretKey: "",
	},
}
