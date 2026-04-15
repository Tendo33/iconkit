package cmd

import (
	"testing"
)

func TestParseSizes_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"16,32,64", []int{16, 32, 64}},
		{"128", []int{128}},
		{"16, 32, 64", []int{16, 32, 64}},
		{" 16 , 32 ", []int{16, 32}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSizes(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("len = %d, want %d", len(result), len(tt.expected))
			}
			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("result[%d] = %d, want %d", i, result[i], v)
				}
			}
		})
	}
}

func TestParseSizes_Invalid(t *testing.T) {
	invalid := []string{
		"abc",
		"16,abc,32",
		"",
		",,,",
		"0",
		"-1",
		"16,-5",
	}

	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			_, err := parseSizes(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}
