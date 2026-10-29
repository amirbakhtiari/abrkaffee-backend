package main

import (
	"log"

	"github.com/amirbakhtiari/abrcoffee/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Printf("AbrCoffee Server is running on port %s...", app.Port)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
