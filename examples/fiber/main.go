// Command fiber is a demo backend (modular MVC over fiber) that mounts archview.
// It shows that the generic router extractor reads fiber's app.Get/Post (and
// Group prefixes) the same way it reads gin and echo.
package main

import (
	"log"

	"github.com/eularixs/archview"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"archview-example-fiber/internal/user/controller"
	"archview-example-fiber/internal/user/repository"
	"archview-example-fiber/internal/user/service"
)

func main() {
	app := fiber.New()

	userCtl := controller.New(service.New(repository.New()))
	api := app.Group("/api")
	api.Get("/users", userCtl.List)
	api.Get("/users/:id", userCtl.Get)
	api.Post("/users", userCtl.Create)

	av, err := archview.New(archview.Options{Root: ".", BasePath: "/graph", ShowPorts: true})
	if err != nil {
		log.Fatal(err)
	}

	// Mount archview on fiber via the net/http adaptor; its routes are excluded
	// from the graph automatically.
	app.Get("/graph", adaptor.HTTPHandler(av.Handler()))
	app.Get("/graph/data", adaptor.HTTPHandler(av.Handler()))

	log.Println("listening on :8099 — open http://localhost:8099/graph")
	log.Fatal(app.Listen(":8099"))
}
