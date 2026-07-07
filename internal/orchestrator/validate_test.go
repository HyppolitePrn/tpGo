package orchestrator

import "testing"

func TestValidateWorkers(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		want      int
		wantError bool
	}{
		{"valid minimum", 1, 1, false},
		{"valid middle", 8, 8, false},
		{"valid maximum", 100, 100, false},
		{"too low", 0, 3, true},
		{"negative", -5, 3, true},
		{"too high", 101, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateWorkers(tt.input)
			if got != tt.want {
				t.Errorf("ValidateWorkers(%d) = %d, want %d", tt.input, got, tt.want)
			}
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateWorkers(%d) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}
