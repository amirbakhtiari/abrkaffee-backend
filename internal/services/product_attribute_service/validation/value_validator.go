package validation

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
		if v.Name == "" {
			errs = append(errs, NewError(
				"values["+string(rune(i))+"].name",
				"required",
				v.Name,
			))
		}

		if v.SortOrder < 0 {
			errs = append(errs, NewError(
				"values["+string(rune(i))+"].sortOrder",
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
