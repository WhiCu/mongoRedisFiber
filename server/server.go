package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"slices"
	"time"

	"github.com/WhiCu/mongoRedisFiber/app"
	"github.com/WhiCu/mongoRedisFiber/config"
	"github.com/WhiCu/mongoRedisFiber/db"
	"github.com/WhiCu/mongoRedisFiber/db/types"
	appInterface "github.com/WhiCu/mongoRedisFiber/server/app"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/redis/go-redis/v9"
)

const (
	RegistrationPage = "/registration"
	HomePage         = "/home"
)

var (
	ProtectedURLs = []string{"/registration", "/db"}
)

type Server struct {
	app    appInterface.AppInterface
	router *fiber.App
}

func NewServer(router *fiber.App, app appInterface.AppInterface) *Server {
	return &Server{
		app:    app,
		router: router,
	}
}

func NewStandardServer(uriDB string) *Server {
	mux := fiber.New(fiber.Config{
		AppName:           "main_app",
		Prefork:           true,
		CaseSensitive:     true,
		EnablePrintRoutes: true,
		StrictRouting:     false,
	})

	//TODO: переработой оформление, должно вcё быть доступно по конфигу, а также оформи

	cont, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	//TODO: mongoDB
	client := db.Client(cont, uriDB)
	if client == nil {

		//TODO: add logging
		log.Fatal("client is nil")
	}
	cancel()

	db := db.NewDB(client, config.MustGet("MONGO_DB"))

	//TODO: redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(config.DefaultGet("REDIS_HOST", "localhost"), config.MustGet("REDIS_PORT")),
		Password: config.MustGet("REDIS_PASSWORD"),
		DB:       config.MustGetInt("REDIS_DB_ID"),
	})

	cont, cancel = context.WithTimeout(context.Background(), 5*time.Second)

	if err := redisClient.Ping(cont).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		panic(err)
	}
	cancel()

	app := app.NewApp(db, redisClient)

	sv := NewServer(mux, app)

	sv.StandartMiddleware()
	sv.StandartRoutes()

	return sv
}

func (sv *Server) StandartMiddleware() {

	sv.router.Use(func(c *fiber.Ctx) error {
		//TODO: add logging
		fmt.Println(c.IP(), c.Path())
		return c.Next()
	})

	sv.router.Use(keyauth.New(keyauth.Config{
		Next: func(c *fiber.Ctx) bool {
			return slices.Contains(ProtectedURLs, c.Path())
		},
		KeyLookup: "cookie:token",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			if sv.app.CorrectToken(key) {
				return true, nil
			}
			return false, keyauth.ErrMissingOrMalformedAPIKey

		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Redirect(RegistrationPage)
		},
	}))

}

func (sv *Server) StandartRoutes() {

	sv.router.Post("/mouse", func(c *fiber.Ctx) error {
		return c.SendString("(·˔·^ )∫ " + string(c.Request().Body()) + " " + c.Method())
	})
	sv.router.Get("/mouse", func(c *fiber.Ctx) error {
		return c.SendString("(·˔·^ )∫")
	})

	sv.router.Post("/db", func(c *fiber.Ctx) error {
		var user types.User
		if err := c.BodyParser(&user); err != nil {
			return err
		}
		types.Token(&user)

		token := sv.app.CheckOrAddUser(&user)

		c.Cookie(&fiber.Cookie{
			Name:  "token",
			Value: token,
		})

		return c.SendStatus(fiber.StatusOK)
	})
	sv.router.Get(RegistrationPage, func(c *fiber.Ctx) error {
		//TODO: доделай страницу
		return c.SendFile("./server/static/registration.html")
	})

	sv.router.All("/*", func(c *fiber.Ctx) error {
		return c.Redirect(RegistrationPage)
	})

}

func (s *Server) Run() error {
	return s.router.Listen(net.JoinHostPort(config.DefaultGet("SERVER_HOST", "localhost"), config.MustGet("SERVER_PORT")))
}
