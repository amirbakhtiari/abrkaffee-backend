package validation

import (
	"strconv"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports/product_attribute"
)

func ValidateCreateAttribute(dto product_attribute.CreateAttributeDTO) error {
	var errs Errors

	if dto.Name == "" {
		errs = append(errs, NewError("name", "required", dto.Name))
	}

	if dto.SortOrder < 0 {
		errs = append(errs, NewError("sortOrder", "must be >= 0", dto.SortOrder))
	}

	if len(dto.Values) == 0 {
		errs = append(errs, NewError("values", "at least one value required", nil))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func ValidateUpdateAttribute(dto product_attribute.UpdateAttributeDTO) error {
	var errs Errors

	if dto.Name == "" {
		errs = append(errs, NewError("name", "required", dto.Name))
	}

	if dto.SortOrder < 0 {
		errs = append(errs, NewError("sortOrder", "must be >= 0", dto.SortOrder))
	}

	for i, v := range dto.Values {
		if v.Name == "" {
			errs = append(errs, NewError("values["+strconv.Itoa(i)+"].name", "required", v.Name))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
