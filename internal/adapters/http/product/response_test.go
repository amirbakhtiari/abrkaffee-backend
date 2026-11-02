package product

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

func TestPublicResponseDoesNotExposeInternalIdentifiers(t *testing.T) {
	dto := &ports.ProductResponse{
		ID: 1, Name: "Americano", Code: "AM-01",
		SKU: "INTERNAL-SKU", Barcode: "123456789",
		Description: "Coffee", IsActive: true,
		CategoryIDs: []uint64{2},
	}

	response := ToPublicResponse(dto)
	if response.Code != "AM-01" {
		t.Fatalf("public code = %q, want AM-01", response.Code)
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal public response: %v", err)
	}
	body := string(encoded)
	if strings.Contains(body, "sku") || strings.Contains(body, "barcode") {
		t.Fatalf("public response leaked internal fields: %s", body)
	}
}
