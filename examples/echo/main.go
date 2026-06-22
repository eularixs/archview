// Command echo is a demo backend (modular MVC over echo) that mounts archview.
// It shows how archview is used with echo: the route extractor reads the
// e.GET/POST/... registrations (including group prefixes), and archview itself
// is served on its own port — the cleanest way to mount it with any framework.
package main

import (
	"log"
	"net/http"

	"github.com/eularixs/archview"
	"github.com/labstack/echo/v4"

	"archview-example-echo/internal/user/controller"
	"archview-example-echo/internal/user/repository"
	"archview-example-echo/internal/user/service"
)

func main() {
	e := echo.New()

	userCtl := controller.New(service.New(repository.New()))
	api := e.Group("/api")
	api.GET("/users", userCtl.List)
	api.GET("/users/:id", userCtl.Get)
	api.POST("/users", userCtl.Create)

	av, err := archview.New(archview.Options{Root: ".", BasePath: "/graph", ShowPorts: true})
	if err != nil {
		log.Fatal(err)
	}

	// Option A (used here): serve archview on its own port — framework-agnostic.
	go func() {
		mux := http.NewServeMux()
		av.Mount(mux)
		log.Println("archview on :9098 — open http://localhost:9098/graph")
		log.Fatal(http.ListenAndServe(":9098", mux))
	}()

	// Option B (in-process): mount on echo directly, e.g.
	//   e.GET("/graph", echo.WrapHandler(av.Handler()))
	//   e.GET("/graph/data", echo.WrapHandler(av.Handler()))

	log.Println("echo on :8098")
	e.Logger.Fatal(e.Start(":8098"))
}
