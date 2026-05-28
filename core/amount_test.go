package core_test

import (
	"testing"

	"github.com/egermano/balde/core"
)

func TestParseAmount_US(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1000", 100000},
		{"1000.00", 100000},
		{"1,234.56", 123456},
		{"0.99", 99},
		{"-50.00", -5000},
		{"-1,000.50", -100050},
		{"0", 0},
	}

	for _, tt := range tests {
		got, err := core.ParseAmount(tt.input, ".", ",")
		if err != nil {
			t.Errorf("ParseAmount(%q, '.', ','): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("ParseAmount(%q, '.', ',') = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestParseAmount_BR(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1000", 100000},
		{"1000,00", 100000},
		{"1.234,56", 123456},
		{"0,99", 99},
		{"-50,00", -5000},
		{"-1.000,50", -100050},
	}

	for _, tt := range tests {
		got, err := core.ParseAmount(tt.input, ",", ".")
		if err != nil {
			t.Errorf("ParseAmount(%q, ',', '.'): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("ParseAmount(%q, ',', '.') = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestParseAmount_Invalid(t *testing.T) {
	_, err := core.ParseAmount("abc", ".", ",")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestFormatAmount_US(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{100000, "$1,000.00"},
		{123456, "$1,234.56"},
		{99, "$0.99"},
		{0, "$0.00"},
		{-5000, "-$50.00"},
	}

	for _, tt := range tests {
		got := core.FormatAmount(tt.cents, ".", ",", "$")
		if got != tt.expected {
			t.Errorf("FormatAmount(%d, '.', ',', '$') = %q, want %q", tt.cents, got, tt.expected)
		}
	}
}

func TestFormatAmount_BR(t *testing.T) {
	got := core.FormatAmount(123456, ",", ".", "R$")
	if got != "R$1.234,56" {
		t.Errorf("FormatAmount(123456, ',', '.', 'R$') = %q, want %q", got, "R$1.234,56")
	}
}
