package domain

import (
	"testing"
	"time"
)

func TestParseTimeControl(t *testing.T) {
	tc, err := ParseTimeControl("10|5")
	if err != nil {
		t.Fatalf("ParseTimeControl returned error: %v", err)
	}

	if tc.Base != 10*time.Minute {
		t.Fatalf("expected 10 minutes, got %v", tc.Base)
	}
	if tc.Increment != 5*time.Second {
		t.Fatalf("expected 5 seconds, got %v", tc.Increment)
	}
	if tc.String() != "10|5" {
		t.Fatalf("expected canonical time control string, got %q", tc.String())
	}
}

func TestParseTimeControlRejectsInvalidValues(t *testing.T) {
	inputs := []string{"0|0", "10", "10|-1", "rapid"}
	for _, input := range inputs {
		if _, err := ParseTimeControl(input); err == nil {
			t.Fatalf("expected %q to be rejected", input)
		}
	}
}
