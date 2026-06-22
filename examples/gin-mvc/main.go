// Command gin-mvc is a demo backend (modular MVC over gin) that mounts archview
// at /graph. Run it, then open http://localhost:8080/graph.
package main

import (
	"log"
	"net/http"

	"github.com/eularixs/archview"
	"github.com/gin-gonic/gin"

	pc "archview-example-gin-mvc/internal/product/controller"
	pr "archview-example-gin-mvc/internal/product/repository"
	ps "archview-example-gin-mvc/internal/product/service"
	uc "archview-example-gin-mvc/internal/user/controller"
	ur "archview-example-gin-mvc/internal/user/repository"
	us "archview-example-gin-mvc/internal/user/service"
)

func main() {
	r := gin.Default()

	// Wire layers: repository -> service -> controller.
	userCtl := uc.New(us.New(ur.New()))
	productCtl := pc.New(ps.New(pr.New()))

	api := r.Group("/api")
	api.GET("/users", userCtl.List)
	api.GET("/users/:id", userCtl.Get)
	api.POST("/users", userCtl.Create)
	api.GET("/products", productCtl.List)
	api.POST("/products", productCtl.Create)

	// archview analyzes this module's source at startup (dev-live).
	av, err := archview.New(archview.Options{
		Root:     ".",
		BasePath: "/graph",
		Editor:   "vscode",
	})
	if err != nil {
		log.Fatal(err)
	}

	// archview owns /graph; gin handles everything else.
	mux := http.NewServeMux()
	av.Mount(mux)
	mux.Handle("/", r)

	log.Println("listening on :8080 — open http://localhost:8080/graph")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
