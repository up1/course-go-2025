package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"demo/models"

	"github.com/labstack/echo/v4"
)

// In a real app, this would be a database call/service layer.
var mockProducts = map[int]models.Product{
	1: {ID: 1, Name: "Product name 1", Price: 100.50, Stock: 10},
}

// ProductHandler holds dependencies like the logger and service/database connection.
type ProductHandler struct {
	Logger *slog.Logger
}

// GetProductByID
// @Summary Get product detail by ID
// @Description Retrieves a product's details using its ID.
// @ID get-product-by-id
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {object} models.ErrorResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /product/{id} [get]
func (h *ProductHandler) GetProductByID(c echo.Context) error {
	// 1. Get and validate path parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Logger.Error("Invalid product ID format",
			slog.String("id", idStr),
			slog.Any("error", err),
			slog.String("traceID", c.Response().Header().Get("X-Request-ID")))

		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid product ID format",
		})
	}

	// 2. Simulated database lookup
	product, found := mockProducts[id]

	// 3. Handle success (200) or not found (404)
	if !found {
		h.Logger.Warn("Product not found", slog.Int("id", id))
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "product id=" + idStr + " not found in system",
		})
	}

	// Logging success (using the slog logger)
	h.Logger.Info("Product retrieved successfully",
		slog.Int("id", id),
		slog.String("name", product.Name))

	return c.JSON(http.StatusOK, product)
}
