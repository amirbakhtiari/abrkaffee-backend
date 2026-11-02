package product

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func testClock() Clock {
	now := time.Date(2026, time.November, 2, 10, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}

func TestNewProductNormalizesFieldsAndCategoryIDs(t *testing.T) {
	item, err := NewProduct(
		"  Americano  ", "  AM-01 ", "  SKU-01 ", " 123456 ",
		"  Hot coffee  ", true, []uint64{3, 1, 3, 0}, testClock(),
	)
	if err != nil {
		t.Fatalf("NewProduct() error = %v", err)
	}

	if item.Name() != "Americano" || item.Code() != "AM-01" {
		t.Fatalf("unexpected normalized product: name=%q code=%q", item.Name(), item.Code())
	}
	if item.SKU() != "SKU-01" || item.Barcode() != "123456" {
		t.Fatalf("SKU or barcode was not normalized")
	}
	if got := item.CategoryIDs(); len(got) != 2 || got[0] != 3 || got[1] != 1 {
		t.Fatalf("CategoryIDs() = %v, want [3 1]", got)
	}
}

func TestNewProductRejectsInvalidRequiredFields(t *testing.T) {
	_, err := NewProduct(" ", "P-1", "", "", "", true, nil, testClock())
	if !errors.Is(err, ErrProductNameRequired) {
		t.Fatalf("NewProduct() error = %v, want ErrProductNameRequired", err)
	}

	_, err = NewProduct("Coffee", " ", "", "", "", true, nil, testClock())
	if !errors.Is(err, ErrProductCodeRequired) {
		t.Fatalf("NewProduct() error = %v, want ErrProductCodeRequired", err)
	}
}

func TestNewProductRejectsOverlongOptionalFields(t *testing.T) {
	_, err := NewProduct(
		"Coffee", "P-1", strings.Repeat("s", MaxProductSKULength+1),
		"", "", true, nil, testClock(),
	)
	if !errors.Is(err, ErrProductSKUTooLong) {
		t.Fatalf("NewProduct() error = %v, want ErrProductSKUTooLong", err)
	}

	_, err = NewProduct(
		"Coffee", "P-1", "", strings.Repeat("1", MaxProductBarcodeLength+1),
		"", true, nil, testClock(),
	)
	if !errors.Is(err, ErrProductBarcodeTooLong) {
		t.Fatalf("NewProduct() error = %v, want ErrProductBarcodeTooLong", err)
	}
}

func TestEditAndStatusChanges(t *testing.T) {
	item, err := NewProduct("Coffee", "P-1", "", "", "", true, []uint64{1}, testClock())
	if err != nil {
		t.Fatal(err)
	}

	if err := item.Edit("Latte", "P-2", "SKU-2", "9988", "Milk coffee", []uint64{2, 3}, testClock()); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}
	item.Deactivate(testClock())

	if item.Name() != "Latte" || item.Code() != "P-2" || item.IsActive() {
		t.Fatalf("unexpected product state after edit/deactivate")
	}
	if got := item.CategoryIDs(); len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("CategoryIDs() = %v, want [2 3]", got)
	}

	item.Activate(testClock())
	if !item.IsActive() {
		t.Fatal("Activate() did not activate product")
	}
}

func TestCategoryIDsReturnsCopy(t *testing.T) {
	item, err := NewProduct("Coffee", "P-1", "", "", "", true, []uint64{1}, testClock())
	if err != nil {
		t.Fatal(err)
	}

	ids := item.CategoryIDs()
	ids[0] = 99
	if item.CategoryIDs()[0] != 1 {
		t.Fatal("CategoryIDs() exposed internal slice")
	}
}
