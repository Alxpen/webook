package config

type Config struct {
	DB    DBConfig
	Redis RedisConfig
	SMS   SMSConfig
}

type SMSConfig struct {
	AppId     string
	SignName  string
	SecretId  string
	SecretKey string
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}
