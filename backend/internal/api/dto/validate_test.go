package dto

import (
	"errors"
	"testing"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

type testRegisterLike struct {
	Email string `json:"email" validate:"required,email"`
	Title string `json:"title" validate:"required,min=3,max=120"`
	Goal  int64  `json:"goal_amount" validate:"required,gt=0"`
}

func TestValidate_PopulatesDetailsByJSONFieldName(t *testing.T) {
	err := Validate(testRegisterLike{Email: "not-an-email", Title: "ab", Goal: 0})
	if err == nil {
		t.Fatal("expected a validation error")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperr.Error, got %T", err)
	}
	if appErr.Code != "validation_failed" {
		t.Fatalf("expected validation_failed, got %s", appErr.Code)
	}

	for _, field := range []string{"email", "title", "goal_amount"} {
		if _, ok := appErr.Details[field]; !ok {
			t.Errorf("expected details for field %q, got %+v", field, appErr.Details)
		}
	}
}

func TestValidate_NoErrorWhenValid(t *testing.T) {
	err := Validate(testRegisterLike{Email: "ana@example.com", Title: "Campaña válida", Goal: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
