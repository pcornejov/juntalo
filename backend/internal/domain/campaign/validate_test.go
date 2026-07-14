package campaign

import (
	"testing"
	"time"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

func TestValidateType(t *testing.T) {
	if err := ValidateType(TypeCollection); err != nil {
		t.Errorf("collection should be enabled: %v", err)
	}
	if err := ValidateType(TypeRaffle); err != nil {
		t.Errorf("raffle should be enabled: %v", err)
	}
	if err := ValidateType(TypeCourse); !apperr.Is(err, "campaign_type_disabled") {
		t.Errorf("course should still be disabled, got %v", err)
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

func TestValidatePublishAt(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	if err := ValidatePublishAt(nil, StatusDraft); err != nil {
		t.Errorf("nil publish_at should not error: %v", err)
	}
	if err := ValidatePublishAt(&future, StatusDraft); err != nil {
		t.Errorf("future publish_at on draft should not error: %v", err)
	}
	if err := ValidatePublishAt(&past, StatusDraft); !apperr.Is(err, "validation_failed") {
		t.Errorf("past publish_at should error, got %v", err)
	}
	if err := ValidatePublishAt(&future, StatusActive); !apperr.Is(err, "campaign_not_active") {
		t.Errorf("publish_at on non-draft should error, got %v", err)
	}
}

func TestValidateRaffleSettings(t *testing.T) {
	price := money.CLP(1000)
	total := 100
	zero := 0

	if err := ValidateRaffleSettings(nil, &total); err == nil {
		t.Error("nil unit price should error")
	}
	if err := ValidateRaffleSettings(&price, nil); err == nil {
		t.Error("nil total numbers should error")
	}
	if err := ValidateRaffleSettings(&price, &zero); err == nil {
		t.Error("zero total numbers should error")
	}
	tooMany := maxRaffleTotalNumbers + 1
	if err := ValidateRaffleSettings(&price, &tooMany); err == nil {
		t.Error("total numbers above the cap should error")
	}
	if err := ValidateRaffleSettings(&price, &total); err != nil {
		t.Errorf("valid settings should not error: %v", err)
	}
}

func TestValidateRaffleLocked(t *testing.T) {
	if err := ValidateRaffleLocked(StatusDraft); err != nil {
		t.Errorf("draft should not be locked: %v", err)
	}
	if err := ValidateRaffleLocked(StatusActive); !apperr.Is(err, "raffle_locked") {
		t.Errorf("active should be locked, got %v", err)
	}
}

func TestValidateRaffleWinningNumber(t *testing.T) {
	total := 10
	inRange := 5
	outOfRange := 11
	zero := 0

	if err := ValidateRaffleWinningNumber(nil, &total); err != nil {
		t.Errorf("nil winning number should not error: %v", err)
	}
	if err := ValidateRaffleWinningNumber(&inRange, &total); err != nil {
		t.Errorf("in-range winning number should not error: %v", err)
	}
	if err := ValidateRaffleWinningNumber(&outOfRange, &total); err == nil {
		t.Error("out-of-range winning number should error")
	}
	if err := ValidateRaffleWinningNumber(&zero, &total); err == nil {
		t.Error("winning number below 1 should error")
	}
	if err := ValidateRaffleWinningNumber(&inRange, nil); err == nil {
		t.Error("winning number without a total range should error")
	}
}
