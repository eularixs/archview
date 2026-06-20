// Package controller exposes product HTTP handlers.
package controller

import (
	"net/http"

	"archview-example-gin-mvc/internal/product/service"

	"github.com/gin-gonic/gin"
)

// ProductController serves product endpoints over a ProductService.
type ProductController struct {
	svc service.ProductService
}

// New builds a ProductController.
func New(svc service.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

// List handles GET /products.
func (c *ProductController) List(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.svc.ListProducts())
}

// Create handles POST /products.
func (c *ProductController) Create(ctx *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		return
	}
	ctx.JSON(http.StatusCreated, c.svc.CreateProduct(body.Name))
}
