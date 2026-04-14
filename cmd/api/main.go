package main

import (
	"log"

	_ "sipi/docs"
	"sipi/internal/app"
)

// @title Sipi API
// @version 1.0
// @description Backend service for intelligent meeting planning.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("bootstrap application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("run application: %v", err)
	}
}
