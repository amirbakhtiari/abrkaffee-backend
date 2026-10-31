package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	port "github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
	"github.com/gin-gonic/gin"
)

type mockUserService struct {
	createFn func(
		context.Context,
		string,
		string,
		string,
		string,
		*time.Time,
	) (*domain.User, error)

	getByIDFn func(context.Context, uint64) (*domain.User, error)

	listFn func(
		context.Context,
		paginator.Request,
	) ([]*domain.User, int64, error)

	updateFn func(
		context.Context,
		uint64,
		port.UpdateUserDTO,
	) (*domain.User, error)

	deleteFn func(context.Context, uint64) error
}

var _ port.Service = (*mockUserService)(nil)

func (m *mockUserService) Create(
	ctx context.Context,
	name, family, mobile, email string,
	birthDate *time.Time,
) (*domain.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, name, family, mobile, email, birthDate)
	}
	panic("unexpected call: Create")
}

func (m *mockUserService) GetByID(
	ctx context.Context,
	id uint64,
) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	panic("unexpected call: GetByID")
}

func (m *mockUserService) List(
	ctx context.Context,
	req paginator.Request,
) ([]*domain.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, req)
	}
	panic("unexpected call: List")
}

func (m *mockUserService) Update(
	ctx context.Context,
	id uint64,
	dto port.UpdateUserDTO,
) (*domain.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, dto)
	}
	panic("unexpected call: Update")
}

func (m *mockUserService) Delete(
	ctx context.Context,
	id uint64,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	panic("unexpected call: Delete")
}

func newTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.POST("/users", handler.CreateUser)
	r.GET("/users", handler.ListUsers)
	r.GET("/users/:id", handler.GetUserByID)
	r.PUT("/users/:id", handler.UpdateUser)
	r.DELETE("/users/:id", handler.DeleteUser)

	return r
}

func performRequest(
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

func testUser() *domain.User {
	now := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	birthDate := time.Date(1995, time.May, 20, 0, 0, 0, 0, time.UTC)

	return domain.Rehydrate(
		42,
		"Amir",
		"Bakhtiari",
		"amir@example.com",
		"09123456789",
		&birthDate,
		true,
		now,
		now,
	)
}
