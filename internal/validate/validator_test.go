package validate

import (
	"testing"

	playgroundvalidator "github.com/go-playground/validator/v10"
)

type testPayload struct {
	Name string `validate:"required"`
}

func TestStructValidatorValidate(t *testing.T) {
	v := &StructValidator{Validator: playgroundvalidator.New()}

	if err := v.Validate(testPayload{}); err == nil {
		t.Fatal("expected validation error for missing required field")
	}
	if err := v.Validate(testPayload{Name: "ok"}); err != nil {
		t.Fatalf("expected validation to pass, got %v", err)
	}
}
