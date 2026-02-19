package frontmatter

import (
	"testing"
)

func TestValidateFrontmatter_AllRequiredPresent(t *testing.T) {
	fm := map[string]interface{}{
		"title":  "Test",
		"status": "draft",
	}
	schema := SchemaDefinition{
		Required: []string{"title", "status"},
	}

	result := ValidateFrontmatter(fm, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
}

func TestValidateFrontmatter_MissingRequired(t *testing.T) {
	fm := map[string]interface{}{
		"title": "Test",
	}
	schema := SchemaDefinition{
		Required: []string{"title", "status", "uid"},
	}

	result := ValidateFrontmatter(fm, schema)
	if !result.HasErrors() {
		t.Fatal("expected errors for missing required fields")
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(result.Errors), result.Errors)
	}

	fields := map[string]bool{}
	for _, e := range result.Errors {
		fields[e.Field] = true
	}
	if !fields["status"] {
		t.Fatal("expected error for missing 'status'")
	}
	if !fields["uid"] {
		t.Fatal("expected error for missing 'uid'")
	}
}

func TestValidateFrontmatter_AllowedValues_Valid(t *testing.T) {
	fm := map[string]interface{}{
		"status": "draft",
	}
	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"status": {Allowed: []string{"draft", "published", "archived"}},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
}

func TestValidateFrontmatter_AllowedValues_Invalid(t *testing.T) {
	fm := map[string]interface{}{
		"status": "unknown",
	}
	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"status": {Allowed: []string{"draft", "published"}},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if !result.HasErrors() {
		t.Fatal("expected error for disallowed value")
	}
	if result.Errors[0].Field != "status" {
		t.Fatalf("expected error on field 'status', got %q", result.Errors[0].Field)
	}
}

func TestValidateFrontmatter_Pattern_Valid(t *testing.T) {
	fm := map[string]interface{}{
		"slug": "my-valid-slug",
	}
	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"slug": {Pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}
}

func TestValidateFrontmatter_Pattern_Invalid(t *testing.T) {
	fm := map[string]interface{}{
		"slug": "INVALID SLUG!",
	}
	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"slug": {Pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if !result.HasErrors() {
		t.Fatal("expected error for pattern mismatch")
	}
}

func TestValidateFrontmatter_MinMaxLength(t *testing.T) {
	min := 3
	max := 10

	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"name": {MinLength: &min, MaxLength: &max},
		},
	}

	tooShort := map[string]interface{}{"name": "ab"}
	result := ValidateFrontmatter(tooShort, schema)
	if !result.HasErrors() {
		t.Fatal("expected error for too-short value")
	}

	tooLong := map[string]interface{}{"name": "this is way too long"}
	result = ValidateFrontmatter(tooLong, schema)
	if !result.HasErrors() {
		t.Fatal("expected error for too-long value")
	}

	justRight := map[string]interface{}{"name": "hello"}
	result = ValidateFrontmatter(justRight, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors, got %v", result.Errors)
	}

	exactMin := map[string]interface{}{"name": "abc"}
	result = ValidateFrontmatter(exactMin, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors at exact min, got %v", result.Errors)
	}

	exactMax := map[string]interface{}{"name": "abcdefghij"}
	result = ValidateFrontmatter(exactMax, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors at exact max, got %v", result.Errors)
	}
}

func TestValidateFrontmatter_NilFrontmatter(t *testing.T) {
	schema := SchemaDefinition{
		Required: []string{"title", "status"},
	}

	result := ValidateFrontmatter(nil, schema)
	if !result.HasErrors() {
		t.Fatal("expected errors for nil frontmatter with required fields")
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result.Errors))
	}
}

func TestValidateFrontmatter_NonStringValue(t *testing.T) {
	fm := map[string]interface{}{
		"count": 42,
	}
	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"count": {Allowed: []string{"one", "two"}},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if !result.HasErrors() {
		t.Fatal("expected error for non-string value")
	}
	if result.Errors[0].Message != "must be a string" {
		t.Fatalf("expected 'must be a string' message, got %q", result.Errors[0].Message)
	}
}

func TestValidateFrontmatter_FieldNotPresentSkipsValidation(t *testing.T) {
	fm := map[string]interface{}{
		"title": "Test",
	}
	schema := SchemaDefinition{
		Validators: map[string]ValidatorConfig{
			"status": {Allowed: []string{"draft"}},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if result.HasErrors() {
		t.Fatalf("expected no errors when validated field is absent, got %v", result.Errors)
	}
}

func TestValidationError_Error(t *testing.T) {
	e := ValidationError{Field: "title", Message: "missing required field"}
	if e.Error() != "title: missing required field" {
		t.Fatalf("unexpected error string: %q", e.Error())
	}

	e2 := ValidationError{Message: "general error"}
	if e2.Error() != "general error" {
		t.Fatalf("unexpected error string for empty field: %q", e2.Error())
	}

	e3 := ValidationError{Field: "slug", Pointer: "frontmatter.slug", Message: "value does not match pattern"}
	if e3.Error() != "frontmatter.slug: value does not match pattern" {
		t.Fatalf("expected pointer-based error string, got %q", e3.Error())
	}
}

func TestValidateFrontmatter_PointerPopulated(t *testing.T) {
	fm := map[string]interface{}{
		"status": "invalid",
	}
	schema := SchemaDefinition{
		Required: []string{"title"},
		Validators: map[string]ValidatorConfig{
			"status": {Allowed: []string{"draft", "published"}},
		},
	}

	result := ValidateFrontmatter(fm, schema)
	if !result.HasErrors() {
		t.Fatal("expected errors")
	}

	for _, e := range result.Errors {
		if e.Pointer == "" {
			t.Fatalf("expected Pointer to be populated for field %q, got empty", e.Field)
		}
		if e.Pointer != "frontmatter."+e.Field {
			t.Fatalf("expected Pointer 'frontmatter.%s', got %q", e.Field, e.Pointer)
		}
	}
}
