package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

func TestDeleteUser_Success(t *testing.T) {
	var receivedID uint64

	service := &mockUserService{
		deleteFn: func(_ context.Context, id uint64) error {
			receivedID = id
			return nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodDelete, "/users/42", "")

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if receivedID != 42 {
		t.Fatalf("expected ID 42, got %d", receivedID)
	}

	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty response body, got %q", recorder.Body.String())
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))

	for _, id := range []string{"abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			recorder := performRequest(router, http.MethodDelete, "/users/"+id, "")

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return domain.ErrUserNotFound
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodDelete, "/users/42", "")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "user not found") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestDeleteUser_InvalidDomainID(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return domain.ErrInvalidID
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodDelete, "/users/42", "")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return errors.New("database unavailable")
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(router, http.MethodDelete, "/users/42", "")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
