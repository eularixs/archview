// Package controller exposes user HTTP handlers over echo.
package controller

import (
	"net/http"
	"strconv"

	"archview-example-echo/internal/user/service"

	"github.com/labstack/echo/v4"
)

// UserController serves user endpoints over a UserService.
type UserController struct {
	svc service.UserService
}

// New builds a UserController.
func New(svc service.UserService) *UserController { return &UserController{svc: svc} }

// List handles GET /api/users.
func (c *UserController) List(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, c.svc.ListUsers())
}

// Get handles GET /api/users/:id.
func (c *UserController) Get(ctx echo.Context) error {
	id, _ := strconv.Atoi(ctx.Param("id"))
	u, ok := c.svc.GetUser(id)
	if !ok {
		return ctx.NoContent(http.StatusNotFound)
	}
	return ctx.JSON(http.StatusOK, u)
}

// Create handles POST /api/users.
func (c *UserController) Create(ctx echo.Context) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.Bind(&body); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}
	return ctx.JSON(http.StatusCreated, c.svc.CreateUser(body.Name))
}
