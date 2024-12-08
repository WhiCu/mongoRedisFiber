package app

import "github.com/WhiCu/mongoRedisFiber/app/types"

type AppInterface interface {
	CorrectToken(token string) bool
	CheckOrAddUser(user types.User) string
}
