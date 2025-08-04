package api

import "testing"

func TestIsValidSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid subdomain with hyphen",
			input:    "crazy-rat",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSubdomain(tt.input)
			if result != tt.expected {
				t.Errorf("isValidSubdomain(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
