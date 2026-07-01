package main

import (
	"fmt"

	"github.com/ovenpickled/hop/config"
	"github.com/ovenpickled/hop/handler"
	"github.com/ovenpickled/hop/router"
	"github.com/ovenpickled/hop/store"
)

func main() {
	cfg := config.Load()

	store.InitializeStore(cfg)
	handler.Init(cfg)

	r := router.SetupRouter()

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error: %v", err))
	}
}
