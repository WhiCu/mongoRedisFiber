package main

import (
	"github.com/WhiCu/mongoRedisFiber/config"
	"github.com/WhiCu/mongoRedisFiber/server"
)

func main() {
	uri := "mongodb://" + config.MustGet("MONGO_USER") + ":" + config.MustGet("MONGO_PASSWORD") +
		"@" + config.MustGet("MONGO_HOST") + ":" + config.MustGet("MONGO_PORT") + "/" + config.MustGet("MONGO_DB")
	server := server.NewStandardServer(uri)
	if err := server.Run(); err != nil {
		panic(err)
	}
}
