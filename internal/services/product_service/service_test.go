package product_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

type fakeCategoryRepository struct {
	ids map[uint64]*category.Category
}

func (r *fakeCategoryRepository) Create(context.Context, *category.Category) (*category.Category, error) {
	panic("unexpected call")
}
func (r *fakeCategoryRepository) GetByID(_ context.Context, id uint64) (*category.Category, error) {
	item, ok := r.ids[id]
	if !ok {
		return nil, category.ErrCategoryNotFound
	}
	return item, nil
}
func (r *fakeCategoryRepository) GetBySlug(context.Context, string) (*category.Category, error) {
	panic("unexpected call")
}
func (r *fakeCategoryRepository) GetAll(context.Context, uint64, uint64) ([]*category.Category, error) {
	panic("unexpected call")
}
func (r *fakeCategoryRepository) Update(context.Context, *category.Category) error {
	panic("unexpected call")
}
func (r *fakeCategoryRepository) Delete(context.Context, uint64) error {
	panic("unexpected call")
}

type fakeProductRepository struct {
	item *product.Product

	statusID        uint64
	statusIsActive  bool
	statusUpdatedAt time.Time
	statusErr       error
	statusCalls     int
}

func (r *fakeProductRepository) Create(_ context.Context, item *product.Product) (*product.Product, error) {
	r.item = item
	return item, nil
}
func (r *fakeProductRepository) GetByID(context.Context, uint64) (*product.Product, error) {
	return r.item, nil
}
func (r *fakeProductRepository) List(context.Context, ports.ProductListQuery) ([]*product.Product, int64, error) {
	return []*product.Product{r.item}, 1, nil
}
func (r *fakeProductRepository) Update(_ context.Context, item *product.Product) (*product.Product, error) {
	r.item = item
	return item, nil
}
func (r *fakeProductRepository) Delete(context.Context, uint64) error { return nil }

func TestCreateProductValidatesCategoryReferences(t *testing.T) {
	now := time.Date(2026, time.November, 2, 10, 0, 0, 0, time.UTC)
	cat := category.ReconstructCategory(7, "Drinks", "drinks", 0, true, now, now)
	categories := &fakeCategoryRepository{ids: map[uint64]*category.Category{7: cat}}
	repository := &fakeProductRepository{}
	service := NewProductService(repository, categories, func() time.Time { return now })

	_, err := service.CreateProduct(context.Background(), &ports.CreateProductDTO{
		Name: "Americano", Code: "AM-01", CategoryIDs: []uint64{7, 999},
	})
	if !errors.Is(err, product.ErrProductCategoryInvalid) {
		t.Fatalf("CreateProduct() error = %v, want ErrProductCategoryInvalid", err)
	}
	if repository.item != nil {
		t.Fatal("repository Create() was called for an invalid category")
	}
}

func TestCreateProductPersistsAndReturnsCategoryIDs(t *testing.T) {
	now := time.Date(2026, time.November, 2, 10, 0, 0, 0, time.UTC)
	cat := category.ReconstructCategory(7, "Drinks", "drinks", 0, true, now, now)
	categories := &fakeCategoryRepository{ids: map[uint64]*category.Category{7: cat}}
	repository := &fakeProductRepository{}
	service := NewProductService(repository, categories, func() time.Time { return now })

	result, err := service.CreateProduct(context.Background(), &ports.CreateProductDTO{
		Name: "Americano", Code: "AM-01", SKU: "SKU-01",
		CategoryIDs: []uint64{7, 7},
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	if result == nil || result.Code != "AM-01" {
		t.Fatalf("unexpected product response: %#v", result)
	}
	if len(result.CategoryIDs) != 1 || result.CategoryIDs[0] != 7 {
		t.Fatalf("CategoryIDs = %v, want [7]", result.CategoryIDs)
	}
}

func (r *fakeProductRepository) UpdateStatus(
	_ context.Context,
	id uint64,
	isActive bool,
	updatedAt time.Time,
) error {
	r.statusCalls++
	r.statusID = id
	r.statusIsActive = isActive
	r.statusUpdatedAt = updatedAt

	return r.statusErr
}

func TestUpdateProductStatus(t *testing.T) {
	now := time.Date(2026, time.November, 7, 10, 0, 0, 0, time.UTC)

	repository := &fakeProductRepository{}
	service := NewProductService(
		repository,
		&fakeCategoryRepository{},
		func() time.Time { return now },
	)

	err := service.UpdateProductStatus(context.Background(), 42, false)
	if err != nil {
		t.Fatalf("UpdateProductStatus() error = %v", err)
	}

	if repository.statusCalls != 1 {
		t.Fatalf("UpdateStatus() calls = %d, want 1", repository.statusCalls)
	}

	if repository.statusID != 42 {
		t.Errorf("statusID = %d, want 42", repository.statusID)
	}

	if repository.statusIsActive {
		t.Error("statusIsActive = true, want false")
	}

	if !repository.statusUpdatedAt.Equal(now) {
		t.Errorf(
			"statusUpdatedAt = %v, want %v",
			repository.statusUpdatedAt,
			now,
		)
	}
}

func TestUpdateProductStatusRejectsZeroID(t *testing.T) {
	repository := &fakeProductRepository{}
	service := NewProductService(
		repository,
		&fakeCategoryRepository{},
		time.Now,
	)

	err := service.UpdateProductStatus(context.Background(), 0, true)

	if !errors.Is(err, product.ErrProductNotFound) {
		t.Fatalf("UpdateProductStatus() error = %v, want ErrProductNotFound", err)
	}

	if repository.statusCalls != 0 {
		t.Fatal("repository UpdateStatus() should not be called")
	}
}
