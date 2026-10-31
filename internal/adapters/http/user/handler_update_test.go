package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	port "github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
)

func TestUpdateUser_Success(t *testing.T) {
	var receivedID uint64
	var receivedDTO port.UpdateUserDTO

	service := &mockUserService{
		updateFn: func(
			_ context.Context,
			id uint64,
			dto port.UpdateUserDTO,
		) (*domain.User, error) {
			receivedID = id
			receivedDTO = dto
			return testUser(), nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPut,
		"/users/42",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789",
			"birth_date":"1995-05-20"
		}`,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if receivedID != 42 {
		t.Fatalf("expected ID 42, got %d", receivedID)
	}

	if receivedDTO.Name != "Amir" ||
		receivedDTO.Family != "Bakhtiari" ||
		receivedDTO.Email != "amir@example.com" ||
		receivedDTO.Mobile != "09123456789" {
		t.Fatalf("unexpected DTO: %+v", receivedDTO)
	}

	if receivedDTO.BirthDate == nil ||
		receivedDTO.BirthDate.Format(time.DateOnly) != "1995-05-20" {
		t.Fatalf("unexpected birth date: %v", receivedDTO.BirthDate)
	}
}

func TestUpdateUser_InvalidID(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))

	for _, id := range []string{"abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			recorder := performRequest(
				router,
				http.MethodPut,
				"/users/"+id,
				`{
					"name":"Amir",
					"family":"Bakhtiari",
					"email":"amir@example.com",
					"mobile":"09123456789"
				}`,
			)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}
}

func TestUpdateUser_InvalidJSON(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))
	recorder := performRequest(router, http.MethodPut, "/users/42", `{invalid`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "invalid request body") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestUpdateUser_InvalidBirthDate(t *testing.T) {
	router := newTestRouter(NewHandler(&mockUserService{}))
	recorder := performRequest(
		router,
		http.MethodPut,
		"/users/42",
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

	if !strings.Contains(recorder.Body.String(), "birth_date must use YYYY-MM-DD format") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestUpdateUser_EmptyBirthDate(t *testing.T) {
	var receivedDTO port.UpdateUserDTO

	service := &mockUserService{
		updateFn: func(
			_ context.Context,
			_ uint64,
			dto port.UpdateUserDTO,
		) (*domain.User, error) {
			receivedDTO = dto
			return testUser(), nil
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPut,
		"/users/42",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789",
			"birth_date":""
		}`,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if receivedDTO.BirthDate != nil {
		t.Fatalf("expected nil birth date, got %v", receivedDTO.BirthDate)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	service := &mockUserService{
		updateFn: func(
			_ context.Context,
			_ uint64,
			_ port.UpdateUserDTO,
		) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPut,
		"/users/42",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789"
		}`,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestUpdateUser_InvalidDomainData(t *testing.T) {
	service := &mockUserService{
		updateFn: func(
			_ context.Context,
			_ uint64,
			_ port.UpdateUserDTO,
		) (*domain.User, error) {
			return nil, domain.ErrInvalidMobileRequired
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPut,
		"/users/42",
		`{
			"name":"Amir",
			"family":"Bakhtiari",
			"email":"amir@example.com",
			"mobile":"09123456789"
		}`,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateUser_ServiceError(t *testing.T) {
	service := &mockUserService{
		updateFn: func(
			_ context.Context,
			_ uint64,
			_ port.UpdateUserDTO,
		) (*domain.User, error) {
			return nil, errors.New("database unavailable")
		},
	}

	router := newTestRouter(NewHandler(service))
	recorder := performRequest(
		router,
		http.MethodPut,
		"/users/42",
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
}
