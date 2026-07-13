package campaign

import (
	"testing"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

func TestValidateType(t *testing.T) {
	if err := ValidateType(TypeCollection); err != nil {
		t.Errorf("collection should be enabled: %v", err)
	}
	if err := ValidateType(TypeRaffle); !apperr.Is(err, "campaign_type_disabled") {
		t.Errorf("raffle should be disabled, got %v", err)
	}
	if err := ValidateType("unknown"); !apperr.Is(err, "unknown_campaign_type") {
		t.Errorf("unknown type should error, got %v", err)
	}
}

func TestValidateGoalUpdate(t *testing.T) {
	raised := money.CLP(100_000)
	lower := money.CLP(50_000)
	higher := money.CLP(200_000)

	if err := ValidateGoalUpdate(&lower, raised); !apperr.Is(err, "goal_below_raised") {
		t.Errorf("goal below raised should error, got %v", err)
	}
	if err := ValidateGoalUpdate(&higher, raised); err != nil {
		t.Errorf("goal above raised should not error: %v", err)
	}
	if err := ValidateGoalUpdate(nil, raised); err != nil {
		t.Errorf("nil goal (no meta) should not error: %v", err)
	}
}
