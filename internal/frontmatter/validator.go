package frontmatter

import (
	"fmt"
	"regexp"
)

// ValidationError represents a validation failure for a specific field.
type ValidationError struct {
	Field   string
	Message string
	Pointer string
}

func (e ValidationError) Error() string {
	if e.Pointer != "" {
		return fmt.Sprintf("%s: %s", e.Pointer, e.Message)
	}
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult captures validation issues for a frontmatter map.
type ValidationResult struct {
	Errors []ValidationError
}

func (r *ValidationResult) Add(field, message string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (r ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ValidateFrontmatter validates a frontmatter map against a schema definition.
func ValidateFrontmatter(frontmatter map[string]interface{}, schema SchemaDefinition) ValidationResult {
	result := ValidationResult{}

	if frontmatter == nil {
		frontmatter = map[string]interface{}{}
	}

	for _, required := range schema.Required {
		if _, ok := frontmatter[required]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Field:   required,
				Pointer: "frontmatter." + required,
				Message: "missing required field",
			})
		}
	}

	for field, rules := range schema.Validators {
		value, ok := frontmatter[field]
		if !ok {
			continue
		}
		for _, ve := range validateField(field, value, rules) {
			ve.Pointer = "frontmatter." + field
			result.Errors = append(result.Errors, ve)
		}
	}

	return result
}

func validateField(field string, value interface{}, rules ValidatorConfig) []ValidationError {
	var errs []ValidationError
	str, ok := value.(string)
	if !ok {
		errs = append(errs, ValidationError{
			Field:   field,
			Message: "must be a string",
		})
		return errs
	}

	if len(rules.Allowed) > 0 && !stringInSlice(str, rules.Allowed) {
		errs = append(errs, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("value %q not in allowed set", str),
		})
	}

	if rules.Pattern != "" {
		re, err := regexp.Compile(rules.Pattern)
		if err != nil {
			errs = append(errs, ValidationError{
				Field:   field,
				Message: "invalid validation pattern",
			})
		} else if !re.MatchString(str) {
			errs = append(errs, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("value %q does not match pattern", str),
			})
		}
	}

	if rules.MinLength != nil && len(str) < *rules.MinLength {
		errs = append(errs, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("length must be >= %d", *rules.MinLength),
		})
	}

	if rules.MaxLength != nil && len(str) > *rules.MaxLength {
		errs = append(errs, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("length must be <= %d", *rules.MaxLength),
		})
	}

	if rules.Custom != "" {
		errs = append(errs, ValidationError{
			Field:   field,
			Message: "custom validators not implemented",
		})
	}

	return errs
}

func stringInSlice(value string, candidates []string) bool {
	for _, c := range candidates {
		if c == value {
			return true
		}
	}
	return false
}
