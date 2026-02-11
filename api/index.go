package handler

import (
	"net/http"

	// "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

type SearchInput struct {
	Query string `json:"query" validate:"required,min=2"`
}

func index(c fiber.Ctx) error {
	return c.JSON(&fiber.Map{
		"message": "Welcome to the s2-leads-api!",
	})
}

func favicon(c fiber.Ctx) error {
	return c.SendFile("./public/favicon.ico")
}

// HTTP Handler which Vercel looks for
func Handler(w http.ResponseWriter, r *http.Request) {
	app := fiber.New()

	// CORS Setup
	app.Use(cors.New())

	// Define Routes
	app.Get("/", index)
	app.Get("/favicon.ico", favicon)

	// Serve HTTP
	http.HandlerFunc(adaptor.FiberApp(app)).ServeHTTP(w, r)
}