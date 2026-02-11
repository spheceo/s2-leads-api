package handler

import (
	"net/http"
	"s2-leads-api/lib"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

var (
	validate    = validator.New()
	fiberApp    = buildApp()
	httpHandler = http.HandlerFunc(adaptor.FiberApp(fiberApp))
)

type SearchInput struct {
	BusinessType string `json:"business_type" validate:"required,min=2"`
	City         string `json:"city" validate:"required,min=2"`
	CountryCode  string `json:"country_code" validate:"required,min=2"`
	Limit        int64  `json:"limit" validate:"required,gte=1,lte=500"`
}

func index(c fiber.Ctx) error {
	return c.JSON(&fiber.Map{
		"message": "Welcome to the s2-leads-api!",
	})
}

func search(c fiber.Ctx) error {
	// Accepting body input & validating
	var body SearchInput

	if err := c.Bind().Body(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid JSON body",
		})
	}

	if err := validate.Struct(body); err != nil {
		return c.Status(422).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch coordinates
	coordinates, coordStatus, err := lib.GetCoordinates(body.City, body.CountryCode)
	if err != nil {
		return c.Status(coordStatus).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if len(coordinates) == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "no coordinates found for given city/country",
		})
	}

	// Fetch leads & return
	leads, leadsStatus, err := lib.GetLeads(
		coordinates[0].Lat, coordinates[0].Lon, body.BusinessType, body.CountryCode, body.Limit,
	)
	if err != nil {
		return c.Status(leadsStatus).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(leadsStatus).JSON(leads)
}

func favicon(c fiber.Ctx) error {
	return c.SendFile("./public/favicon.ico")
}

func buildApp() *fiber.App {
	app := fiber.New()

	// CORS Setup
	app.Use(cors.New())

	// Define Routes
	app.Get("/", index)
	app.Post("/search", unkeyAuth, search)
	app.Get("/favicon.ico", favicon)

	return app
}

// HTTP Handler which Vercel looks for
func Handler(w http.ResponseWriter, r *http.Request) {
	httpHandler.ServeHTTP(w, r)
}
