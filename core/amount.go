package core

import (
	"fmt"
	"strings"
)

func ParseAmount(input string, decimalSep, thousandsSep string) (int64, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}

	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	s = strings.ReplaceAll(s, thousandsSep, "")

	parts := strings.Split(s, decimalSep)
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount: %s", input)
	}

	intPart := parts[0]
	var centsStr string
	if len(parts) == 2 {
		centsStr = parts[1]
		if len(centsStr) > 2 {
			return 0, fmt.Errorf("invalid amount: too many decimal places")
		}
	}

	var intVal int64
	if intPart != "" {
		for _, c := range intPart {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("invalid amount: %s", input)
			}
			intVal = intVal*10 + int64(c-'0')
		}
	}

	var centsVal int64
	if centsStr != "" {
		for _, c := range centsStr {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("invalid amount: %s", input)
			}
			centsVal = centsVal*10 + int64(c-'0')
		}
		if len(centsStr) == 1 {
			centsVal *= 10
		}
	}

	result := intVal*100 + centsVal
	if negative {
		result = -result
	}
	return result, nil
}

func FormatAmount(cents int64, decimalSep, thousandsSep, symbol string) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}

	whole := cents / 100
	frac := cents % 100

	wholeStr := formatWithThousands(whole, thousandsSep)
	result := symbol + wholeStr + decimalSep + fmt.Sprintf("%02d", frac)

	if negative {
		result = "-" + result
	}
	return result
}

func formatWithThousands(n int64, sep string) string {
	if n == 0 {
		return "0"
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	var result []byte
	count := 0
	for i := len(digits) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result = append([]byte(sep), result...)
		}
		result = append([]byte{digits[i]}, result...)
		count++
	}

	return string(result)
}
