package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

func TestCreateUser_Success(t *testing.T) {
	expected := testUser()
	var receivedBirthDate *time.Time

	service := &mockUserService{
		createFn: func(
			_ context.Context,
			name, family, mobile, email string,
			birthDate *time.Time,
		) (*domain.User, error) {
			if name != "Amir" ||
				family != "Bakhtiari" ||
				mobile != "09123456789" ||
				email != "amir@example.com" {
				t.Fatalf("unexpected Create arguments")
			}

			receivedBirthDate = birthDate
			return expected, nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789",
			"birth_date":"1995-05-20"
		}`,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s",
			http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	if receivedBirthDate == nil {
		t.Fatal("expected birth date to be passed to service")
	}

	if receivedBirthDate.Format(time.DateOnly) != "1995-05-20" {
		t.Fatalf("unexpected birth date: %s", receivedBirthDate.Format(time.DateOnly))
	}

	if !strings.Contains(recorder.Body.String(), `"id":42`) {
		t.Fatalf("response does not contain expected user: %s", recorder.Body.String())
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))
	recorder := performRequest(router, http.MethodPost, "/users", `{invalid`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestCreateUser_MissingRequiredField(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"email":"amir@example.com",
			"mobile":"09123456789"
		}`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"not-an-email",
			"mobile":"09123456789"
		}`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestCreateUser_InvalidBirthDate(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789",
			"birth_date":"20-05-1995"
		}`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "invalid birth_date") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestCreateUser_DomainValidationError(t *testing.T) {
	service := &mockUserService{
		createFn: func(
			_ context.Context,
			_, _, _, _ string,
			_ *time.Time,
		) (*domain.User, error) {
			return nil, domain.ErrInvalidMobileRequired
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"bad"
		}`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateUser_ServiceError(t *testing.T) {
	service := &mockUserService{
		createFn: func(
			_ context.Context,
			_, _, _, _ string,
			_ *time.Time,
		) (*domain.User, error) {
			return nil, errors.New("database unavailable")
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789"
		}`,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "failed to create user") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestCreateUser_EmptyBirthDate(t *testing.T) {
	var receivedBirthDate *time.Time

	service := &mockUserService{
		createFn: func(
			_ context.Context,
			_, _, _, _ string,
			birthDate *time.Time,
		) (*domain.User, error) {
			receivedBirthDate = birthDate
			return testUser(), nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPost,
		"/users",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789",
			"birth_date":""
		}`,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if receivedBirthDate != nil {
		t.Fatal("expected nil birth date for empty input")
	}
}
