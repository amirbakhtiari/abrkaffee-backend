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

	recorder := performRequest(
		router,
		http.MethodDelete,
		"/users/42",
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if receivedID != 42 {
		t.Fatalf("expected ID 42, got %d", receivedID)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"کاربر با موفقیت حذف شد.",
	) {
		t.Fatalf(
			"expected success message, got: %s",
			recorder.Body.String(),
		)
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))

	for _, id := range []string{"abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			recorder := performRequest(
				router,
				http.MethodDelete,
				"/users/"+id,
				"",
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

func TestDeleteUser_NotFound(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return domain.ErrUserNotFound
		},
	}

	router := newTestRouter(NewHandler(service))

	recorder := performRequest(
		router,
		http.MethodDelete,
		"/users/42",
		"",
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"کاربر موردنظر پیدا نشد.",
	) {
		t.Fatalf(
			"expected not-found message, got: %s",
			recorder.Body.String(),
		)
	}
}

func TestDeleteUser_InvalidDomainID(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return domain.ErrInvalidID
		},
	}

	router := newTestRouter(NewHandler(service))

	recorder := performRequest(
		router,
		http.MethodDelete,
		"/users/42",
		"",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestDeleteUser_HasAccount(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return domain.ErrUserHasAccount
		},
	}

	router := newTestRouter(NewHandler(service))

	recorder := performRequest(
		router,
		http.MethodDelete,
		"/users/42",
		"",
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected 409, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"امکان حذف آن وجود ندارد",
	) {
		t.Fatalf(
			"expected account restriction message, got: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"غیرفعال کنید",
	) {
		t.Fatalf(
			"expected deactivate guidance, got: %s",
			recorder.Body.String(),
		)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	service := &mockUserService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return errors.New("database unavailable")
		},
	}

	router := newTestRouter(NewHandler(service))

	recorder := performRequest(
		router,
		http.MethodDelete,
		"/users/42",
		"",
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"خطایی هنگام حذف کاربر رخ داد.",
	) {
		t.Fatalf(
			"expected internal error message, got: %s",
			recorder.Body.String(),
		)
	}
}
