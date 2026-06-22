// Command gin-mvc is a demo backend (modular MVC over gin) that mounts archview
// at /graph. Run it, then open http://localhost:8080/graph.
package main

import (
	"log"

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

	// Mount archview on gin itself; its own routes are excluded from the graph.
	r.GET("/graph", gin.WrapH(av.Handler()))
	r.GET("/graph/data", gin.WrapH(av.Handler()))

	log.Println("listening on :8080 — open http://localhost:8080/graph")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
