package main

import (
	"testing"
	"time"
)

func TestMonthsUntilVacation(t *testing.T) {
	now := time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		date string
		want int
		ok   bool
	}{
		{name: "next month", date: "2026-11", want: 1, ok: true},
		{name: "next year", date: "2027-03", want: 5, ok: true},
		{name: "same month rejected", date: "2026-10", ok: false},
		{name: "past month rejected", date: "2026-09", ok: false},
		{name: "invalid date rejected", date: "March 2027", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := monthsUntilVacation(now, tt.date)
			if (err == nil) != tt.ok {
				t.Fatalf("monthsUntilVacation() error = %v, want success=%v", err, tt.ok)
			}
			if tt.ok && got != tt.want {
				t.Errorf("monthsUntilVacation() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMonthlyInstallmentRoundsUp(t *testing.T) {
	tests := []struct {
		budget int64
		months int
		want   int64
	}{
		{budget: 1_200_000, months: 5, want: 240_000},
		{budget: 1_000, months: 3, want: 334},
		{budget: 500, months: 1, want: 500},
	}

	for _, tt := range tests {
		if got := monthlyInstallment(tt.budget, tt.months); got != tt.want {
			t.Errorf("monthlyInstallment(%d, %d) = %d, want %d", tt.budget, tt.months, got, tt.want)
		}
	}
}

func TestVacationBucketName(t *testing.T) {
	tests := []struct {
		label       string
		destination string
		date        string
		want        string
	}{
		{label: "simple", destination: "Japan", date: "2027-03", want: "vacation-japan-2027-03"},
		{label: "spaces and punctuation", destination: "Rio de Janeiro!", date: "2027-03", want: "vacation-rio-de-janeiro-2027-03"},
		{label: "accented", destination: "São Paulo", date: "2027-03", want: "vacation-sao-paulo-2027-03"},
	}

	for _, tt := range tests {
		if got := vacationBucketName(tt.destination, tt.date); got != tt.want {
			t.Errorf("vacationBucketName(%q, %q) = %q, want %q", tt.destination, tt.date, got, tt.want)
		}
	}
}
