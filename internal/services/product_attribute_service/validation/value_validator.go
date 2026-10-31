package validation

import (
	"strconv"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports/product_attribute"
)

func ValidateCreateValue(dto product_attribute.CreateAttributeValueDTO) error {
	var errs Errors

	if dto.Name == "" {
		errs = append(errs, NewError("name", "required", dto.Name))
	}

	if dto.SortOrder < 0 {
		errs = append(errs, NewError("sortOrder", "must be >= 0", dto.SortOrder))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func ValidateValues(dtos []product_attribute.CreateAttributeValueDTO) error {
	var errs Errors

	for i, v := range dtos {
		index := strconv.Itoa(i)

		if v.Name == "" {
			errs = append(errs, NewError(
				"values["+index+"].name",
				"required",
				v.Name,
			))
		}

		if v.SortOrder < 0 {
			errs = append(errs, NewError(
				"values["+index+"].sortOrder",
				"must be >= 0",
				v.SortOrder,
			))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
