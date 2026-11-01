package category

import (
	"errors"
	"testing"
	"time"
)

func fixedClock() Clock {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	return func() time.Time { return now }
}

func TestNewCategoryNormalizesSlug(t *testing.T) {
	cat, err := NewCategory("  Drinks  ", "  HOT-DRINKS  ", true, 2, fixedClock())
	if err != nil {
		t.Fatalf("NewCategory() error = %v", err)
	}

	if cat.Name() != "Drinks" {
		t.Errorf("Name() = %q, want Drinks", cat.Name())
	}
	if cat.Slug() != "hot-drinks" {
		t.Errorf("Slug() = %q, want hot-drinks", cat.Slug())
	}
}

func TestNewCategoryRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name     string
		categoryName string
		slug     string
		priority int
		wantErr  error
	}{
		{name: "empty name", categoryName: " ", slug: "valid", wantErr: ErrCategoryNameEmpty},
		{name: "long name", categoryName: string(make([]byte, MaxCategoryNameLength+1)), slug: "valid", wantErr: ErrCategoryNameTooLong},
		{name: "invalid slug", categoryName: "Drinks", slug: "not valid", wantErr: ErrCategorySlugInvalid},
		{name: "negative priority", categoryName: "Drinks", slug: "drinks", priority: -1, wantErr: ErrCategoryPriorityInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCategory(tt.categoryName, tt.slug, true, tt.priority, fixedClock())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewCategory() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditNormalizesSlugAndUpdatesFields(t *testing.T) {
	clock := fixedClock()
	cat, err := NewCategory("Drinks", "drinks", true, 1, clock)
	if err != nil {
		t.Fatal(err)
	}

	if err := cat.Edit(" Hot Drinks ", " HOT-DRINKS ", 3, clock); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	if cat.Name() != "Hot Drinks" || cat.Slug() != "hot-drinks" || cat.Priority() != 3 {
		t.Fatalf("unexpected category after edit: name=%q slug=%q priority=%d", cat.Name(), cat.Slug(), cat.Priority())
	}
}

func TestChangeSlugNormalizesSlug(t *testing.T) {
	cat, err := NewCategory("Drinks", "drinks", true, 0, fixedClock())
	if err != nil {
		t.Fatal(err)
	}

	if err := cat.ChangeSlug(" HOT-DRINKS ", fixedClock()); err != nil {
		t.Fatalf("ChangeSlug() error = %v", err)
	}
	if cat.Slug() != "hot-drinks" {
		t.Errorf("Slug() = %q, want hot-drinks", cat.Slug())
	}
}
