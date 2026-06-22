// Package controller exposes user HTTP handlers over fiber.
package controller

import (
	"strconv"

	"archview-example-fiber/internal/user/service"

	"github.com/gofiber/fiber/v2"
)

// UserController serves user endpoints over a UserService.
type UserController struct {
	svc service.UserService
}

// New builds a UserController.
func New(svc service.UserService) *UserController { return &UserController{svc: svc} }

// List handles GET /api/users.
func (c *UserController) List(ctx *fiber.Ctx) error {
	return ctx.JSON(c.svc.ListUsers())
}

// Get handles GET /api/users/:id.
func (c *UserController) Get(ctx *fiber.Ctx) error {
	id, _ := strconv.Atoi(ctx.Params("id"))
	u, ok := c.svc.GetUser(id)
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}
	return ctx.JSON(u)
}

// Create handles POST /api/users.
func (c *UserController) Create(ctx *fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}
	return ctx.Status(fiber.StatusCreated).JSON(c.svc.CreateUser(body.Name))
}
