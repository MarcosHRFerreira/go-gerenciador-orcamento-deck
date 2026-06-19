package budgetimport

import "testing"

func TestParseOptionalNumberShouldAcceptBrazilianThousandsSeparator(t *testing.T) {
	value, err := parseOptionalNumber("6.617.099,00", true)
	if err != nil {
		t.Fatalf("expected number without error, got %v", err)
	}
	if value != 6617099 {
		t.Fatalf("expected parsed value 6617099, got %f", value)
	}
}

func TestParseOptionalNumberShouldKeepDotDecimalFormat(t *testing.T) {
	value, err := parseOptionalNumber("65515.83", true)
	if err != nil {
		t.Fatalf("expected number without error, got %v", err)
	}
	if value != 65515.83 {
		t.Fatalf("expected parsed value 65515.83, got %f", value)
	}
}
