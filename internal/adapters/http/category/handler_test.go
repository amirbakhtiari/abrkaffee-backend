package category

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type mockCategoryService struct {
	createFn  func(context.Context, *ports.CreateCategoryDTO) (*ports.CategoryResponse, error)
	updateFn  func(context.Context, uint64, *ports.UpdateCategoryDTO) (*ports.CategoryResponse, error)
	deleteFn  func(context.Context, uint64) error
	getByIDFn func(context.Context, uint64) (*ports.CategoryResponse, error)
	listFn    func(context.Context, ports.ListCategoriesDTO) ([]*ports.CategoryResponse, error)
}

var _ ports.CategoryService = (*mockCategoryService)(nil)

func (m *mockCategoryService) CreateCategory(
	ctx context.Context,
	req *ports.CreateCategoryDTO,
) (*ports.CategoryResponse, error) {
	if m.createFn == nil {
		panic("unexpected call: CreateCategory")
	}
	return m.createFn(ctx, req)
}

func (m *mockCategoryService) UpdateCategory(
	ctx context.Context,
	id uint64,
	req *ports.UpdateCategoryDTO,
) (*ports.CategoryResponse, error) {
	if m.updateFn == nil {
		panic("unexpected call: UpdateCategory")
	}
	return m.updateFn(ctx, id, req)
}

func (m *mockCategoryService) DeleteCategory(
	ctx context.Context,
	id uint64,
) error {
	if m.deleteFn == nil {
		panic("unexpected call: DeleteCategory")
	}
	return m.deleteFn(ctx, id)
}

func (m *mockCategoryService) GetCategoryByID(
	ctx context.Context,
	id uint64,
) (*ports.CategoryResponse, error) {
	if m.getByIDFn == nil {
		panic("unexpected call: GetCategoryByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockCategoryService) ListCategories(
	ctx context.Context,
	dto ports.ListCategoriesDTO,
) ([]*ports.CategoryResponse, error) {
	if m.listFn == nil {
		panic("unexpected call: ListCategories")
	}
	return m.listFn(ctx, dto)
}

func newCategoryTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.POST("/categories", handler.CreateCategory)
	router.GET("/categories", handler.ListCategories)
	router.GET("/categories/:id", handler.GetCategoryByID)
	router.PUT("/categories/:id", handler.UpdateCategory)
	router.DELETE("/categories/:id", handler.DeleteCategory)

	return router
}

func performCategoryRequest(
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

func testCategoryResponse() *ports.CategoryResponse {
	return &ports.CategoryResponse{
		ID:       42,
		Name:     "Coffee",
		Slug:     "coffee",
		Priority: 1,
		IsActive: true,
	}
}

func TestCreateCategory_Success(t *testing.T) {
	service := &mockCategoryService{
		createFn: func(
			_ context.Context,
			req *ports.CreateCategoryDTO,
		) (*ports.CategoryResponse, error) {
			if req.Name != "Coffee" {
				t.Fatalf("unexpected name: %s", req.Name)
			}
			if req.Slug != "coffee" {
				t.Fatalf("unexpected slug: %s", req.Slug)
			}
			if req.Priority != 1 || !req.IsActive {
				t.Fatalf("unexpected category attributes")
			}
			return testCategoryResponse(), nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodPost,
		"/categories",
		`{"name":"Coffee","slug":"coffee","priority":1,"is_active":true}`,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if !strings.Contains(recorder.Body.String(), `"id":42`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestCreateCategory_InvalidJSON(t *testing.T) {
	router := newCategoryTestRouter(NewHandler(&mockCategoryService{}))

	recorder := performCategoryRequest(
		router,
		http.MethodPost,
		"/categories",
		`{invalid`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestCreateCategory_DomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "empty name",
			err:        domain.ErrCategoryNameEmpty,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "name too long",
			err:        domain.ErrCategoryNameTooLong,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid slug",
			err:        domain.ErrCategorySlugInvalid,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid priority",
			err:        domain.ErrCategoryPriorityInvalid,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "duplicate name",
			err:        domain.ErrNameCategoryAlreadyExists,
			statusCode: http.StatusConflict,
		},
		{
			name:       "duplicate slug",
			err:        domain.ErrSlugCategoryAlreadyExists,
			statusCode: http.StatusConflict,
		},
		{
			name:       "internal error",
			err:        errors.New("database unavailable"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockCategoryService{
				createFn: func(
					_ context.Context,
					_ *ports.CreateCategoryDTO,
				) (*ports.CategoryResponse, error) {
					return nil, tt.err
				},
			}

			router := newCategoryTestRouter(NewHandler(service))

			recorder := performCategoryRequest(
				router,
				http.MethodPost,
				"/categories",
				`{"name":"Coffee","slug":"coffee"}`,
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

func TestCreateCategory_NilServiceResult(t *testing.T) {
	service := &mockCategoryService{
		createFn: func(
			_ context.Context,
			_ *ports.CreateCategoryDTO,
		) (*ports.CategoryResponse, error) {
			return nil, nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodPost,
		"/categories",
		`{"name":"Coffee","slug":"coffee"}`,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	service := &mockCategoryService{
		updateFn: func(
			_ context.Context,
			id uint64,
			req *ports.UpdateCategoryDTO,
		) (*ports.CategoryResponse, error) {
			if id != 42 {
				t.Fatalf("expected id 42, got %d", id)
			}
			if req.Name != "Tea" || req.Slug != "tea" {
				t.Fatalf("unexpected update request: %+v", req)
			}
			if req.IsActive == nil || !*req.IsActive {
				t.Fatal("expected is_active to be true")
			}
			return testCategoryResponse(), nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodPut,
		"/categories/42",
		`{"name":"Tea","slug":"tea","priority":2,"is_active":true}`,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateCategory_InvalidID(t *testing.T) {
	router := newCategoryTestRouter(NewHandler(&mockCategoryService{}))

	tests := []string{"abc", "0", "-1"}

	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			recorder := performCategoryRequest(
				router,
				http.MethodPut,
				"/categories/"+id,
				`{"name":"Tea","slug":"tea","priority":1,"is_active":true}`,
			)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}
}

func TestUpdateCategory_InvalidJSON(t *testing.T) {
	router := newCategoryTestRouter(NewHandler(&mockCategoryService{}))

	recorder := performCategoryRequest(
		router,
		http.MethodPut,
		"/categories/42",
		`{invalid`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	service := &mockCategoryService{
		updateFn: func(
			_ context.Context,
			_ uint64,
			_ *ports.UpdateCategoryDTO,
		) (*ports.CategoryResponse, error) {
			return nil, domain.ErrCategoryNotFound
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodPut,
		"/categories/42",
		`{"name":"Tea","slug":"tea","priority":1,"is_active":true}`,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestGetCategoryByID_Success(t *testing.T) {
	service := &mockCategoryService{
		getByIDFn: func(
			_ context.Context,
			id uint64,
		) (*ports.CategoryResponse, error) {
			if id != 42 {
				t.Fatalf("unexpected id: %d", id)
			}
			return testCategoryResponse(), nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories/42",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	service := &mockCategoryService{
		getByIDFn: func(
			_ context.Context,
			_ uint64,
		) (*ports.CategoryResponse, error) {
			return nil, domain.ErrCategoryNotFound
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories/42",
		"",
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestGetCategoryByID_InvalidID(t *testing.T) {
	router := newCategoryTestRouter(NewHandler(&mockCategoryService{}))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories/0",
		"",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestListCategories_Success(t *testing.T) {
	service := &mockCategoryService{
		listFn: func(
			_ context.Context,
			dto ports.ListCategoriesDTO,
		) ([]*ports.CategoryResponse, error) {
			if dto.Skip != 5 || dto.Take != 10 {
				t.Fatalf("unexpected pagination: %+v", dto)
			}
			return []*ports.CategoryResponse{
				testCategoryResponse(),
			}, nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories?skip=5&take=10",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if !strings.Contains(recorder.Body.String(), `"id":42`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestListCategories_ServiceError(t *testing.T) {
	service := &mockCategoryService{
		listFn: func(
			_ context.Context,
			_ ports.ListCategoriesDTO,
		) ([]*ports.CategoryResponse, error) {
			return nil, errors.New("database unavailable")
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories",
		"",
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}

	if strings.Contains(recorder.Body.String(), "database unavailable") {
		t.Fatalf("internal error leaked: %s", recorder.Body.String())
	}
}

func TestListCategories_NilItem(t *testing.T) {
	service := &mockCategoryService{
		listFn: func(
			_ context.Context,
			_ ports.ListCategoriesDTO,
		) ([]*ports.CategoryResponse, error) {
			return []*ports.CategoryResponse{nil}, nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories",
		"",
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestListCategories_InvalidQuery(t *testing.T) {
	router := newCategoryTestRouter(NewHandler(&mockCategoryService{}))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories?skip=invalid",
		"",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	service := &mockCategoryService{
		deleteFn: func(
			_ context.Context,
			id uint64,
		) error {
			if id != 42 {
				t.Fatalf("unexpected id: %d", id)
			}
			return nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodDelete,
		"/categories/42",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	service := &mockCategoryService{
		deleteFn: func(
			_ context.Context,
			_ uint64,
		) error {
			return domain.ErrCategoryNotFound
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodDelete,
		"/categories/42",
		"",
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestDeleteCategory_InvalidID(t *testing.T) {
	router := newCategoryTestRouter(NewHandler(&mockCategoryService{}))

	recorder := performCategoryRequest(
		router,
		http.MethodDelete,
		"/categories/0",
		"",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestListCategories_DefaultPagination(t *testing.T) {
	service := &mockCategoryService{
		listFn: func(
			_ context.Context,
			dto ports.ListCategoriesDTO,
		) ([]*ports.CategoryResponse, error) {
			if dto.Skip != 0 {
				t.Errorf("expected skip 0, got %d", dto.Skip)
			}

			if dto.Take != 10 {
				t.Errorf("expected take 10, got %d", dto.Take)
			}

			return []*ports.CategoryResponse{}, nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if body := strings.TrimSpace(recorder.Body.String()); body != "[]" {
		t.Fatalf("expected empty JSON array, got %s", body)
	}
}

func TestUpdateCategory_NilServiceResult(t *testing.T) {
	service := &mockCategoryService{
		updateFn: func(
			_ context.Context,
			id uint64,
			req *ports.UpdateCategoryDTO,
		) (*ports.CategoryResponse, error) {
			if id != 42 {
				t.Errorf("expected id 42, got %d", id)
			}

			if req == nil {
				t.Fatal("expected non-nil update request")
			}

			return nil, nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodPut,
		"/categories/42",
		`{
			"name": "Tea",
			"slug": "tea",
			"priority": 2,
			"is_active": true
		}`,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(recorder.Body.String(), "internal server error") {
		t.Fatalf("expected generic internal error, got %s", recorder.Body.String())
	}
}

func TestGetCategoryByID_NilServiceResult(t *testing.T) {
	service := &mockCategoryService{
		getByIDFn: func(
			_ context.Context,
			id uint64,
		) (*ports.CategoryResponse, error) {
			if id != 42 {
				t.Errorf("expected id 42, got %d", id)
			}

			return nil, nil
		},
	}

	router := newCategoryTestRouter(NewHandler(service))

	recorder := performCategoryRequest(
		router,
		http.MethodGet,
		"/categories/42",
		"",
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(recorder.Body.String(), "internal server error") {
		t.Fatalf("expected generic internal error, got %s", recorder.Body.String())
	}
}
