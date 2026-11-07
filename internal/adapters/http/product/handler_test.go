package product

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type mockProductService struct {
	updateStatusFn func(context.Context, uint64, bool) error
}

var _ ports.ProductService = (*mockProductService)(nil)

func (m *mockProductService) CreateProduct(
	context.Context,
	*ports.CreateProductDTO,
) (*ports.ProductResponse, error) {
	panic("unexpected call: CreateProduct")
}

func (m *mockProductService) UpdateProduct(
	context.Context,
	uint64,
	*ports.UpdateProductDTO,
) (*ports.ProductResponse, error) {
	panic("unexpected call: UpdateProduct")
}

func (m *mockProductService) DeleteProduct(
	context.Context,
	uint64,
) error {
	panic("unexpected call: DeleteProduct")
}

func (m *mockProductService) GetProductByID(
	context.Context,
	uint64,
) (*ports.ProductResponse, error) {
	panic("unexpected call: GetProductByID")
}

func (m *mockProductService) ListProducts(
	context.Context,
	ports.ListProductsDTO,
) ([]*ports.ProductResponse, int64, error) {
	panic("unexpected call: ListProducts")
}

func (m *mockProductService) UpdateProductStatus(
	ctx context.Context,
	id uint64,
	isActive bool,
) error {
	if m.updateStatusFn == nil {
		panic("unexpected call: UpdateProductStatus")
	}

	return m.updateStatusFn(ctx, id, isActive)
}

func newStatusTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.PATCH("/products/:id/status", handler.UpdateProductStatus)

	return router
}

func performStatusRequest(
	router http.Handler,
	method string,
	target string,
	body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func TestUpdateProductStatus_Success(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{
			name:     "activate product",
			body:     `{"is_active":true}`,
			expected: true,
		},
		{
			name:     "deactivate product",
			body:     `{"is_active":false}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false

			service := &mockProductService{
				updateStatusFn: func(
					_ context.Context,
					id uint64,
					isActive bool,
				) error {
					called = true

					if id != 42 {
						t.Errorf("expected id 42, got %d", id)
					}

					if isActive != tt.expected {
						t.Errorf(
							"expected isActive %v, got %v",
							tt.expected,
							isActive,
						)
					}

					return nil
				},
			}

			router := newStatusTestRouter(NewHandler(service))

			recorder := performStatusRequest(
				router,
				http.MethodPatch,
				"/products/42/status",
				tt.body,
			)

			if recorder.Code != http.StatusOK {
				t.Fatalf(
					"expected 200, got %d: %s",
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if !called {
				t.Fatal("expected UpdateProductStatus to be called")
			}

			if !strings.Contains(
				recorder.Body.String(),
				"product status updated successfully",
			) {
				t.Fatalf("unexpected response: %s", recorder.Body.String())
			}
		})
	}
}

func TestUpdateProductStatus_IdempotentRequest(t *testing.T) {
	calls := 0

	service := &mockProductService{
		updateStatusFn: func(
			_ context.Context,
			id uint64,
			isActive bool,
		) error {
			calls++

			if id != 42 {
				t.Errorf("expected id 42, got %d", id)
			}

			if isActive {
				t.Error("expected isActive to be false")
			}

			return nil
		},
	}

	router := newStatusTestRouter(NewHandler(service))

	for i := 0; i < 2; i++ {
		recorder := performStatusRequest(
			router,
			http.MethodPatch,
			"/products/42/status",
			`{"is_active":false}`,
		)

		if recorder.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, recorder.Code)
		}
	}

	if calls != 2 {
		t.Fatalf("expected 2 service calls, got %d", calls)
	}
}

func TestUpdateProductStatus_InvalidRequests(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "invalid product id",
			path: "/products/abc/status",
			body: `{"is_active":true}`,
		},
		{
			name: "zero product id",
			path: "/products/0/status",
			body: `{"is_active":true}`,
		},
		{
			name: "missing is_active",
			path: "/products/42/status",
			body: `{}`,
		},
		{
			name: "null is_active",
			path: "/products/42/status",
			body: `{"is_active":null}`,
		},
		{
			name: "invalid JSON",
			path: "/products/42/status",
			body: `{invalid`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newStatusTestRouter(
				NewHandler(&mockProductService{}),
			)

			recorder := performStatusRequest(
				router,
				http.MethodPatch,
				tt.path,
				tt.body,
			)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected 400, got %d: %s",
					recorder.Code,
					recorder.Body.String(),
				)
			}
		})
	}
}

func TestUpdateProductStatus_DomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "product not found",
			err:        domain.ErrProductNotFound,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "internal error",
			err:        errors.New("database unavailable"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockProductService{
				updateStatusFn: func(
					_ context.Context,
					_ uint64,
					_ bool,
				) error {
					return tt.err
				},
			}

			router := newStatusTestRouter(NewHandler(service))

			recorder := performStatusRequest(
				router,
				http.MethodPatch,
				"/products/42/status",
				`{"is_active":true}`,
			)

			if recorder.Code != tt.statusCode {
				t.Fatalf(
					"expected %d, got %d: %s",
					tt.statusCode,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if tt.statusCode == http.StatusInternalServerError &&
				strings.Contains(recorder.Body.String(), "database unavailable") {
				t.Fatalf("internal error leaked: %s", recorder.Body.String())
			}
		})
	}
}
