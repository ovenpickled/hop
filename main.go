package main

import (
	"fmt"
	"github.com/ovenpickled/hop/router"
	"github.com/ovenpickled/hop/store"
)

func main() {
	store.InitializeStore()

	r := router.SetupRouter()

	err := r.Run(":9808")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error: %v", err))
	}
}
