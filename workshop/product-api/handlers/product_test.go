package handlers_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"demo/handlers"
	"demo/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetProductByID(t *testing.T) {
	e := echo.New()

	// Setup a test handler with a dummy slog logger
	testLogger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	h := &handlers.ProductHandler{Logger: testLogger}

	// Define the test cases
	tests := []struct {
		name         string
		productID    string
		expectedCode int
		expectedBody interface{}
	}{
		{
			name:         "Success: Product Found (ID 1)",
			productID:    "1",
			expectedCode: http.StatusOK,
			expectedBody: models.Product{ID: 1, Name: "Product name 1", Price: 100.50, Stock: 10},
		},
		{
			name:         "Failure: Product Not Found (ID 2)",
			productID:    "2",
			expectedCode: http.StatusNotFound,
			expectedBody: models.ErrorResponse{Message: "product id=2 not found in system"},
		},
		{
			name:         "Failure: Invalid ID Format (abc)",
			productID:    "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: models.ErrorResponse{Message: "Invalid product ID format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup request and recorder
			req := httptest.NewRequest(http.MethodGet, "/product/"+tt.productID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Manually set path/param to simulate Echo router
			c.SetPath("/product/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.productID)

			// The Echo context must be set with the logger to avoid panic in handler
			c.Set("logger", testLogger)

			// Execute the handler
			if assert.NoError(t, h.GetProductByID(c)) {
				assert.Equal(t, tt.expectedCode, rec.Code)

				// Unmarshal the response body into the expected type for comparison
				if tt.expectedCode == http.StatusOK {
					var actualBody models.Product
					if err := json.Unmarshal(rec.Body.Bytes(), &actualBody); err != nil {
						t.Fatalf("Failed to unmarshal response body: %v", err)
					}
					assert.Equal(t, tt.expectedBody, actualBody)
				} else {
					var actualBody models.ErrorResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &actualBody); err != nil {
						t.Fatalf("Failed to unmarshal response body: %v", err)
					}
					assert.Equal(t, tt.expectedBody, actualBody)
				}
			}
		})
	}
}
