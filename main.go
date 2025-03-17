package main

import (
	"example.com/m/config"
	"example.com/m/handlers"
	"github.com/labstack/echo/v4"
)

func main() {
	// Initialize configurations, databases, and other services
	config.Init()

	e := echo.New()

	// Setup routes
	handlers.SetupRoutes(e)

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}
