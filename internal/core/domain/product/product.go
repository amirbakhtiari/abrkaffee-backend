package product

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxProductNameLength        = 255
	MaxProductCodeLength        = 64
	MaxProductSKULength         = 100
	MaxProductBarcodeLength     = 100
	MaxProductDescriptionLength = 5000
)

type Clock func() time.Time

var DefaultClock Clock = time.Now

type Product struct {
	id          uint64
	name        string
	code        string
	sku         string
	barcode     string
	description string
	isActive    bool
	categoryIDs []uint64
	createdAt   time.Time
	updatedAt   time.Time
}

func NewProduct(
	name, code, sku, barcode, description string,
	isActive bool,
	categoryIDs []uint64,
	clock Clock,
) (*Product, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	sku = strings.TrimSpace(sku)
	barcode = strings.TrimSpace(barcode)
	description = strings.TrimSpace(description)

	if err := validateFields(name, code, sku, barcode, description); err != nil {
		return nil, err
	}

	now := clock()
	return &Product{
		name:        name,
		code:        code,
		sku:         sku,
		barcode:     barcode,
		description: description,
		isActive:    isActive,
		categoryIDs: normalizeCategoryIDs(categoryIDs),
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func ReconstructProduct(
	id uint64,
	name, code, sku, barcode, description string,
	isActive bool,
	categoryIDs []uint64,
	createdAt, updatedAt time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		code:        code,
		sku:         sku,
		barcode:     barcode,
		description: description,
		isActive:    isActive,
		categoryIDs: normalizeCategoryIDs(categoryIDs),
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p *Product) ID() uint64             { return p.id }
func (p *Product) Name() string           { return p.name }
func (p *Product) Code() string           { return p.code }
func (p *Product) SKU() string            { return p.sku }
func (p *Product) Barcode() string        { return p.barcode }
func (p *Product) Description() string    { return p.description }
func (p *Product) IsActive() bool         { return p.isActive }
func (p *Product) CreatedAt() time.Time   { return p.createdAt }
func (p *Product) UpdatedAt() time.Time   { return p.updatedAt }
func (p *Product) CategoryIDs() []uint64 {
	ids := make([]uint64, len(p.categoryIDs))
	copy(ids, p.categoryIDs)
	return ids
}

func (p *Product) Edit(
	name, code, sku, barcode, description string,
	categoryIDs []uint64,
	clock Clock,
) error {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	sku = strings.TrimSpace(sku)
	barcode = strings.TrimSpace(barcode)
	description = strings.TrimSpace(description)

	if err := validateFields(name, code, sku, barcode, description); err != nil {
		return err
	}

	p.name = name
	p.code = code
	p.sku = sku
	p.barcode = barcode
	p.description = description
	p.categoryIDs = normalizeCategoryIDs(categoryIDs)
	p.updatedAt = clock()
	return nil
}

func (p *Product) Activate(clock Clock) {
	if !p.isActive {
		p.isActive = true
		p.updatedAt = clock()
	}
}

func (p *Product) Deactivate(clock Clock) {
	if p.isActive {
		p.isActive = false
		p.updatedAt = clock()
	}
}

func validateFields(name, code, sku, barcode, description string) error {
	if name == "" {
		return ErrProductNameRequired
	}
	if utf8.RuneCountInString(name) > MaxProductNameLength {
		return ErrProductNameTooLong
	}
	if code == "" {
		return ErrProductCodeRequired
	}
	if utf8.RuneCountInString(code) > MaxProductCodeLength {
		return ErrProductCodeTooLong
	}
	if utf8.RuneCountInString(sku) > MaxProductSKULength {
		return ErrProductSKUTooLong
	}
	if utf8.RuneCountInString(barcode) > MaxProductBarcodeLength {
		return ErrProductBarcodeTooLong
	}
	if utf8.RuneCountInString(description) > MaxProductDescriptionLength {
		return ErrProductDescriptionTooLong
	}
	return nil
}

func normalizeCategoryIDs(ids []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(ids))
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
