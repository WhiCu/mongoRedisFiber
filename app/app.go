package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/WhiCu/mongoRedisFiber/app/types"
	"github.com/WhiCu/mongoRedisFiber/db"
	dbtypes "github.com/WhiCu/mongoRedisFiber/db/types"
	"github.com/redis/go-redis/v9"
)

type App struct {
	db    *db.DB
	redis *redis.Client
}

func NewApp(db *db.DB, redis *redis.Client) *App {
	return &App{
		db:    db,
		redis: redis,
	}
}

func (app *App) CorrectToken(token string) bool {
	log.Println("token: ", token)

	val, err := app.redis.Get(context.Background(), "user:"+token).Result()
	if err == redis.Nil {
		log.Println("value not found in redis: ", val)
	} else if err != nil {
		log.Fatalf("failed to get value, error: %v\n", err)
	} else {
		log.Println("value found in redis: ", val)
		var user dbtypes.User
		if err := json.Unmarshal([]byte(val), &user); err != nil {
			log.Fatalf("failed to unmarshal value, error: %v\n", err)
		}
		return user.GetToken() == token
	}

	user := app.db.FindToken(context.Background(), "users", token)
	log.Println("value found in db: ", user)
	if user == nil || user.GetToken() == "" {
		return false
	}
	app.AddUserInRedis(user)
	return true
}

func (app *App) AddUserInRedis(user types.User) {

	switch user := user.(type) {
	case *dbtypes.User:
		jsonUser, err := json.Marshal(user)

		if err != nil {
			log.Fatalf("failed to marshal value, error: %v\n", err)
		}

		app.redis.Set(context.Background(), "user:"+user.GetToken(), jsonUser, 30*time.Minute)
	default:
		log.Fatalf("unknown type: %T\n", user)
	}

}

func (app *App) CheckOrAddUser(user types.User) string {

	if app.CorrectToken(user.GetToken()) {
		return user.GetToken()
	}
	app.AddUserInRedis(user)

	dbUser := user.(*dbtypes.User)

	_, token := app.db.AddUser(context.Background(), "users", dbUser)

	return token

}
