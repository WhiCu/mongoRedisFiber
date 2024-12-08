package types

import (
	"crypto/sha256"
	"encoding/base64"
)

type User struct {
	//Id       string `json:"id" bson:"_id"`
	Login    string `json:"login" bson:"login"`
	Password string `json:"password" bson:"password"`
	Token    string `json:"token" bson:"token"`
}

// New создает нового пользователя
func New(login, password string) *User {
	u := &User{Login: login, Password: password}
	Token(u)
	return u
}

// ID возвращает ID пользователя
// func (u *User) ID() string {
// 	return u.Id
// }

// Key возвращает ключ пользователя
func (u *User) Key() string {
	return u.Login + u.Password
}
func (u *User) GetToken() string {
	if u.Token == "" {
		Token(u)
	}
	return u.Token
}

// Token возвращает токен пользователя
func Token(u *User) string {
	hasher := sha256.New()
	hasher.Write([]byte(u.Key()))
	hash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	u.Token = hash
	return hash
}
