package category_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

type categoryRepoStub struct {
	item       *category.Category
	all        []*category.Category
	gotSkip    uint64
	gotTake    uint64
	updateErr  error
	getByIDErr error
}

func (r *categoryRepoStub) Create(_ context.Context, item *category.Category) (*category.Category, error) {
	r.item = item
	return item, nil
}
func (r *categoryRepoStub) GetByID(_ context.Context, _ uint64) (*category.Category, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	return r.item, nil
}
func (r *categoryRepoStub) GetBySlug(context.Context, string) (*category.Category, error) {
	return r.item, nil
}
func (r *categoryRepoStub) GetAll(_ context.Context, skip, take uint64) ([]*category.Category, error) {
	r.gotSkip, r.gotTake = skip, take
	return r.all, nil
}
func (r *categoryRepoStub) Update(_ context.Context, item *category.Category) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.item = item
	return nil
}
func (r *categoryRepoStub) Delete(context.Context, uint64) error { return nil }

func testClock() category.Clock {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	return func() time.Time { return now }
}

func TestUpdateCategoryAppliesActiveStateWhenProvided(t *testing.T) {
	repo := &categoryRepoStub{}
	repo.item, _ = category.NewCategory("Drinks", "drinks", true, 0, testClock())
	svc := NewCategoryService(repo, testClock())

	inactive := false
	got, err := svc.UpdateCategory(context.Background(), 1, &ports.UpdateCategoryDTO{
		Name: "Drinks", Slug: "drinks", Priority: 0, IsActive: &inactive,
	})
	if err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}
	if got.IsActive {
		t.Fatal("expected category to be inactive")
	}
}

func TestUpdateCategoryPreservesActiveStateWhenOmitted(t *testing.T) {
	repo := &categoryRepoStub{}
	repo.item, _ = category.NewCategory("Drinks", "drinks", false, 0, testClock())
	svc := NewCategoryService(repo, testClock())

	got, err := svc.UpdateCategory(context.Background(), 1, &ports.UpdateCategoryDTO{
		Name: "Drinks", Slug: "drinks", Priority: 0,
	})
	if err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}
	if got.IsActive {
		t.Fatal("expected omitted IsActive to preserve inactive state")
	}
}

func TestListCategoriesDefaultsOnlyZeroTake(t *testing.T) {
	repo := &categoryRepoStub{}
	svc := NewCategoryService(repo, testClock())

	_, err := svc.ListCategories(context.Background(), ports.ListCategoriesDTO{Skip: 0, Take: 50})
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if repo.gotTake != 50 {
		t.Fatalf("repository take = %d, want 50", repo.gotTake)
	}

	_, err = svc.ListCategories(context.Background(), ports.ListCategoriesDTO{Skip: 20, Take: 0})
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if repo.gotTake != 10 {
		t.Fatalf("repository take = %d, want default 10", repo.gotTake)
	}
}

func TestCreateCategoryRejectsNilRequest(t *testing.T) {
	svc := NewCategoryService(&categoryRepoStub{}, testClock())
	_, err := svc.CreateCategory(context.Background(), nil)
	if !errors.Is(err, ErrCategoryRequestRequired) {
		t.Fatalf("CreateCategory() error = %v, want %v", err, ErrCategoryRequestRequired)
	}
}

func TestUpdateCategoryRejectsNilRequest(t *testing.T) {
	svc := NewCategoryService(&categoryRepoStub{}, testClock())
	_, err := svc.UpdateCategory(context.Background(), 1, nil)
	if !errors.Is(err, ErrCategoryRequestRequired) {
		t.Fatalf("UpdateCategory() error = %v, want %v", err, ErrCategoryRequestRequired)
	}
}
