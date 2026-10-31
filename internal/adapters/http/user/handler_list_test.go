package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

func TestListUsers_Success(t *testing.T) {
	service := &mockUserService{
		listFn: func(
			_ context.Context,
			req paginator.Request,
		) ([]*domain.User, int64, error) {
			if req.Page != 2 {
				t.Errorf("expected page 2, got %d", req.Page)
			}
			if req.PageSize != 10 {
				t.Errorf("expected page size 10, got %d", req.PageSize)
			}

			return []*domain.User{testUser()}, 21, nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodGet,
		"/users?page=2&page_size=10",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, expected := range []string{
		`"items"`,
		`"id":42`,
		`"page":2`,
		`"pageSize":10`,
		`"totalRow":21`,
		`"totalPage":3`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response missing %s: %s", expected, body)
		}
	}
}

func TestListUsers_InvalidPagination(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))

	tests := []struct {
		name string
		url  string
	}{
		{"missing page", "/users?page_size=10"},
		{"missing page size", "/users?page=1"},
		{"invalid page", "/users?page=abc&page_size=10"},
		{"zero page", "/users?page=0&page_size=10"},
		{"negative page", "/users?page=-1&page_size=10"},
		{"invalid page size", "/users?page=1&page_size=abc"},
		{"zero page size", "/users?page=1&page_size=0"},
		{"negative page size", "/users?page=1&page_size=-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := performRequest(router, http.MethodGet, tt.url, "")

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestListUsers_EmptyResult(t *testing.T) {
	service := &mockUserService{
		listFn: func(
			_ context.Context,
			_ paginator.Request,
		) ([]*domain.User, int64, error) {
			return []*domain.User{}, 0, nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodGet,
		"/users?page=1&page_size=10",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `"totalRow":0`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestListUsers_NilItemsAreSkipped(t *testing.T) {
	service := &mockUserService{
		listFn: func(
			_ context.Context,
			_ paginator.Request,
		) ([]*domain.User, int64, error) {
			return []*domain.User{nil, testUser()}, 2, nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodGet,
		"/users?page=1&page_size=10",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if strings.Count(body, `"id":42`) != 1 {
		t.Fatalf("expected exactly one serialized user: %s", body)
	}
}

func TestListUsers_ServiceError(t *testing.T) {
	service := &mockUserService{
		listFn: func(
			_ context.Context,
			_ paginator.Request,
		) ([]*domain.User, int64, error) {
			return nil, 0, errors.New("database unavailable")
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodGet,
		"/users?page=1&page_size=10",
		"",
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
