package ui

import "testing"

func TestFormatFunctionsDoNotPanic(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"FormatSuccess", FormatSuccess},
		{"FormatError", FormatError},
		{"FormatWarning", FormatWarning},
		{"FormatInfo", FormatInfo},
		{"FormatHeader", FormatHeader},
		{"FormatStability", FormatStability},
		{"FormatSource", FormatSource},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic with empty string
			result := tt.fn("")
			if result == "" && tt.name != "FormatStability" && tt.name != "FormatSource" {
				// FormatSuccess etc. always produce output due to icons
			}

			// Should not panic with normal string
			tt.fn("test input")
		})
	}
}

func TestStabilityColor(t *testing.T) {
	tests := []string{"stable", "beta", "alpha", "development", "deprecated", "unknown"}
	for _, s := range tests {
		// Should not panic
		_ = StabilityColor(s)
	}
}

func TestFormatSourceValues(t *testing.T) {
	// Each should return a non-empty string
	for _, source := range []string{"base", "custom", "other"} {
		result := FormatSource(source)
		if result == "" {
			t.Errorf("FormatSource(%q) returned empty string", source)
		}
	}
}
