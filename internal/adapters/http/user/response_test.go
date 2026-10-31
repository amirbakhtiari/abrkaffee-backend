package user

import (
	"testing"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

func TestToResponse(t *testing.T) {
	u := testUser()
	got := toResponse(u)

	if got.ID != u.ID() {
		t.Errorf("ID: expected %d, got %d", u.ID(), got.ID)
	}
	if got.Name != u.Name() {
		t.Errorf("Name: expected %q, got %q", u.Name(), got.Name)
	}
	if got.Family != u.Family() {
		t.Errorf("Family: expected %q, got %q", u.Family(), got.Family)
	}
	if got.Email != u.Email() {
		t.Errorf("Email: expected %q, got %q", u.Email(), got.Email)
	}
	if got.Mobile != u.Mobile() {
		t.Errorf("Mobile: expected %q, got %q", u.Mobile(), got.Mobile)
	}
	if got.IsActive != u.IsActive() {
		t.Errorf("IsActive: expected %v, got %v", u.IsActive(), got.IsActive)
	}
	if !got.CreatedAt.Equal(u.CreatedAt()) {
		t.Errorf("CreatedAt mismatch")
	}
	if !got.UpdatedAt.Equal(u.UpdatedAt()) {
		t.Errorf("UpdatedAt mismatch")
	}

	if got.BirthDate == nil {
		t.Fatal("expected birth date")
	}

	if got.BirthDate.Format(time.DateOnly) != "1995-05-20" {
		t.Errorf("unexpected birth date: %s", got.BirthDate.Format(time.DateOnly))
	}
}

func TestToResponse_NilBirthDate(t *testing.T) {
	u := testUser()
	u = domain.Rehydrate(
		u.ID(),
		u.Name(),
		u.Family(),
		u.Email(),
		u.Mobile(),
		nil,
		u.IsActive(),
		u.CreatedAt(),
		u.UpdatedAt(),
	)

	got := toResponse(u)

	if got.BirthDate != nil {
		t.Fatalf("expected nil birth date, got %v", got.BirthDate)
	}
}
