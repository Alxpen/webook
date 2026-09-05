package domain

import "time"

// 领域对象， 是 DDD 中的聚合根， 是 DDD 中的 entity
// BO (Business Object)
type User struct {
	Id       int64
	Email    string
	Password string

	Nickname string
	Birthday time.Time
	AboutMe  string
	Ctime    time.Time
}
