package utils

import "testing"

func TestSliceOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []string
		expected bool
	}{
		{
			name:     "overlapping slices",
			a:        []string{"admin", "user"},
			b:        []string{"user", "viewer"},
			expected: true,
		},
		{
			name:     "no overlap",
			a:        []string{"admin"},
			b:        []string{"viewer"},
			expected: false,
		},
		{
			name:     "empty slice",
			a:        []string{},
			b:        []string{"admin"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceOverlaps(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
