package budget

import "testing"

func TestEstimate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"whitespace only", "   \n\t  \n", 0},
		{"one word", "hello", 1},
		{"three words", "alpha beta gamma", 3},
		{"newlines + tabs", "alpha\nbeta\tgamma\nrho", 4},
		{"markdown link", "see [reasoning](corpus/X.md) for context", 4}, // markdown link is one whitespace-token
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Estimate([]byte(tt.in)); got != tt.want {
				t.Errorf("Estimate(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
