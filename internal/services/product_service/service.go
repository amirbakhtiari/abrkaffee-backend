package product_service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

var (
	ErrProductRequestRequired = errors.New("product request is required")
	ErrImageRequired          = errors.New("product image is required")
)

const productImageURLPrefix = "/uploads/"

type productService struct {
	repo         ports.ProductRepository
	categoryRepo ports.CategoryRepository
	storage      ports.FileStorage
	clock        product.Clock
}

func NewProductService(
	repo ports.ProductRepository,
	categoryRepo ports.CategoryRepository,
	storage ports.FileStorage,
	clock product.Clock,
) ports.ProductService {
	return &productService{
		repo:         repo,
		categoryRepo: categoryRepo,
		storage:      storage,
		clock:        clock,
	}
}

func (s *productService) CreateProduct(
	ctx context.Context,
	req *ports.CreateProductDTO,
) (response *ports.ProductResponse, retErr error) {
	if req == nil {
		return nil, ErrProductRequestRequired
	}
	if req.Image == nil {
		return nil, product.ErrProductImageRequired
	}

	if err := s.validateCategories(ctx, req.CategoryIDs); err != nil {
		return nil, err
	}

	imageKey, err := s.storage.Save(ctx, req.Image, req.ImageType)
	if err != nil {
		return nil, fmt.Errorf("save product image: %w", err)
	}

	keepImage := false

	defer func() {
		if keepImage {
			return
		}

		if cleanupErr := s.storage.Delete(
			context.WithoutCancel(ctx),
			imageKey,
		); cleanupErr != nil {
			retErr = errors.Join(
				retErr,
				fmt.Errorf("cleanup product image: %w", cleanupErr),
			)
		}
	}()

	item, err := product.NewProduct(
		req.Name,
		req.Code,
		req.SKU,
		req.Barcode,
		req.Description,
		imageKey,
		req.IsActive,
		req.CategoryIDs,
		s.clock,
	)
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, errors.New("repository returned nil product after create")
	}

	keepImage = true
	return s.toServiceResponse(saved), nil
}

func (s *productService) UpdateProduct(
	ctx context.Context,
	id uint64,
	req *ports.UpdateProductDTO,
) (*ports.ProductResponse, error) {
	if req == nil {
		return nil, ErrProductRequestRequired
	}
	if id == 0 {
		return nil, product.ErrProductNotFound
	}

	if err := s.validateCategories(ctx, req.CategoryIDs); err != nil {
		return nil, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := item.Edit(
		req.Name,
		req.Code,
		req.SKU,
		req.Barcode,
		req.Description,
		req.CategoryIDs,
		s.clock,
	); err != nil {
		return nil, err
	}

	if req.IsActive != nil {
		if *req.IsActive {
			item.Activate(s.clock)
		} else {
			item.Deactivate(s.clock)
		}
	}

	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, err
	}

	return s.toServiceResponse(updated), nil
}

func (s *productService) UpdateProductImage(
	ctx context.Context,
	id uint64,
	file io.Reader,
	contentType string,
) error {
	if id == 0 {
		return product.ErrProductNotFound
	}
	if file == nil {
		return product.ErrProductImageRequired
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return product.ErrProductNotFound
	}

	oldImageKey := item.ImageKey()

	newImageKey, err := s.storage.Save(ctx, file, contentType)
	if err != nil {
		return fmt.Errorf("save product image: %w", err)
	}

	if err := item.ChangeImage(newImageKey, s.clock); err != nil {
		_ = s.storage.Delete(context.WithoutCancel(ctx), newImageKey)
		return err
	}

	if err := s.repo.UpdateImage(
		ctx,
		id,
		newImageKey,
		item.UpdatedAt(),
	); err != nil {
		cleanupErr := s.storage.Delete(
			context.WithoutCancel(ctx),
			newImageKey,
		)
		if cleanupErr != nil {
			return errors.Join(
				fmt.Errorf("update product image in repository: %w", err),
				fmt.Errorf("cleanup new image: %w", cleanupErr),
			)
		}
		return fmt.Errorf("update product image in repository: %w", err)
	}

	if oldImageKey != "" && oldImageKey != newImageKey {
		if err := s.storage.Delete(context.WithoutCancel(ctx), oldImageKey); err != nil {
			return fmt.Errorf(
				"product image updated, but old image cleanup failed: %w",
				err,
			)
		}
	}

	return nil
}

func (s *productService) DeleteProduct(
	ctx context.Context,
	id uint64,
) error {
	if id == 0 {
		return product.ErrProductNotFound
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return product.ErrProductNotFound
	}

	imageKey := item.ImageKey()

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if imageKey != "" {
		if err := s.storage.Delete(context.WithoutCancel(ctx), imageKey); err != nil {
			return fmt.Errorf(
				"product deleted, but image cleanup failed: %w",
				err,
			)
		}
	}

	return nil
}

func (s *productService) GetProductByID(
	ctx context.Context,
	id uint64,
) (*ports.ProductResponse, error) {
	if id == 0 {
		return nil, product.ErrProductNotFound
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toServiceResponse(item), nil
}

func (s *productService) ListProducts(
	ctx context.Context,
	dto ports.ListProductsDTO,
) ([]*ports.ProductResponse, int64, error) {
	if dto.Page < 1 {
		dto.Page = 1
	}
	if dto.PageSize < 1 {
		dto.PageSize = 10
	}
	if dto.PageSize > 100 {
		dto.PageSize = 100
	}

	dto.Search = strings.TrimSpace(dto.Search)

	items, total, err := s.repo.List(ctx, ports.ProductListQuery{
		Page:       dto.Page,
		PageSize:   dto.PageSize,
		Search:     dto.Search,
		IsActive:   dto.IsActive,
		CategoryID: dto.CategoryID,
	})
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*ports.ProductResponse, len(items))
	for i, item := range items {
		responses[i] = s.toServiceResponse(item)
	}

	return responses, total, nil
}

func (s *productService) UpdateProductStatus(
	ctx context.Context,
	id uint64,
	isActive bool,
) error {
	if id == 0 {
		return product.ErrProductNotFound
	}

	return s.repo.UpdateStatus(ctx, id, isActive, s.clock())
}

func (s *productService) validateCategories(
	ctx context.Context,
	ids []uint64,
) error {
	seen := make(map[uint64]struct{}, len(ids))

	for _, id := range ids {
		if id == 0 {
			return product.ErrProductCategoryInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}

		cat, err := s.categoryRepo.GetByID(ctx, id)
		if errors.Is(err, category.ErrCategoryNotFound) ||
			(err == nil && cat == nil) {
			return product.ErrProductCategoryInvalid
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *productService) toServiceResponse(
	item *product.Product,
) *ports.ProductResponse {
	if item == nil {
		return nil
	}

	imageURL := ""
	if item.ImageKey() != "" {
		imageURL = productImageURLPrefix + item.ImageKey()
	}

	return &ports.ProductResponse{
		ID:          item.ID(),
		Name:        item.Name(),
		Code:        item.Code(),
		SKU:         item.SKU(),
		Barcode:     item.Barcode(),
		Description: item.Description(),
		IsActive:    item.IsActive(),
		CategoryIDs: item.CategoryIDs(),
		ImageURL:    imageURL,
		CreatedAt:   item.CreatedAt(),
		UpdatedAt:   item.UpdatedAt(),
	}
}
