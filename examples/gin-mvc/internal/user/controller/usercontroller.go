// Package controller exposes user HTTP handlers.
package controller

import (
	"net/http"
	"strconv"

	"archview-example-gin-mvc/internal/user/service"

	"github.com/gin-gonic/gin"
)

// UserController serves user endpoints over a UserService.
type UserController struct {
	svc service.UserService
}

// New builds a UserController.
func New(svc service.UserService) *UserController {
	return &UserController{svc: svc}
}

// List handles GET /users.
func (c *UserController) List(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.svc.ListUsers())
}

// Get handles GET /users/:id.
func (c *UserController) Get(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	u, ok := c.svc.GetUser(id)
	if !ok {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, u)
}

// Create handles POST /users.
func (c *UserController) Create(ctx *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		return
	}
	ctx.JSON(http.StatusCreated, c.svc.CreateUser(body.Name))
}
