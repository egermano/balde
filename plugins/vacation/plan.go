package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// monthsUntilVacation returns the number of calendar-month allocations from
// now until the vacation month. The current and past months are rejected: a
// plan must have at least one future monthly allocation.
func monthsUntilVacation(now time.Time, date string) (int, error) {
	vacation, err := time.Parse("2006-01", date)
	if err != nil {
		return 0, fmt.Errorf("vacation date must be YYYY-MM: %w", err)
	}
	months := (vacation.Year()-now.Year())*12 + int(vacation.Month()-now.Month())
	if months <= 0 {
		return 0, fmt.Errorf("vacation date %s must be in a future month", date)
	}
	return months, nil
}

// monthlyInstallment rounds upward so the planned target is reached by the
// vacation month even when cents do not divide equally.
func monthlyInstallment(budget int64, months int) int64 {
	return (budget + int64(months) - 1) / int64(months)
}

func vacationBucketName(name, date string) string {
	return "vacation-" + slug(name) + "-" + date
}

func slug(input string) string {
	// Common Latin accents keep names readable without pulling a transliteration
	// dependency into this standalone plugin.
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u", "ç", "c",
		"Á", "a", "À", "a", "Â", "a", "Ã", "a", "Ä", "a",
		"É", "e", "È", "e", "Ê", "e", "Ë", "e",
		"Í", "i", "Ì", "i", "Î", "i", "Ï", "i",
		"Ó", "o", "Ò", "o", "Ô", "o", "Õ", "o", "Ö", "o",
		"Ú", "u", "Ù", "u", "Û", "u", "Ü", "u", "Ç", "c",
	)
	input = replacer.Replace(input)

	var out strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(input) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out.WriteRune(r)
			lastDash = false
			continue
		}
		if out.Len() > 0 && !lastDash {
			out.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(out.String(), "-")
}
