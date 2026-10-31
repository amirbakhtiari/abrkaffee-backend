package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

func TestGetUserByID_Success(t *testing.T) {
	service := &mockUserService{
		getByIDFn: func(_ context.Context, id uint64) (*domain.User, error) {
			if id != 42 {
				t.Fatalf("expected ID 42, got %d", id)
			}
			return testUser(), nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodGet, "/users/42", "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if !strings.Contains(recorder.Body.String(), `"id":42`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))

	for _, id := range []string{"abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			recorder := performRequest(router, http.MethodGet, "/users/"+id, "")

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	service := &mockUserService{
		getByIDFn: func(_ context.Context, _ uint64) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodGet, "/users/42", "")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestGetUserByID_ServiceError(t *testing.T) {
	service := &mockUserService{
		getByIDFn: func(_ context.Context, _ uint64) (*domain.User, error) {
			return nil, errors.New("database unavailable")
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodGet, "/users/42", "")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
