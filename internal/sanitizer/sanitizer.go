package sanitizer

import (
	"errors"
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrEmptyField   = errors.New("field cannot be empty or contain only whitespace")
	ErrInvalidPrice = errors.New("invalid price: must be non-negative decimal or 'free'")
	ErrInvalidStock = errors.New("invalid stock: must be non-negative integer")
	ErrInvalidWeight = errors.New("invalid weight: must be positive decimal")

	scriptRegex = regexp.MustCompile(`(?i)<script[\s\S]*?>[\s\S]*?</script>`)
	tagRegex    = regexp.MustCompile(`(?i)<[^>]+>`)
)

func SanitizeText(input string) string {
	cleaned := scriptRegex.ReplaceAllString(input, "")
	cleaned = tagRegex.ReplaceAllString(cleaned, "")
	cleaned = html.EscapeString(cleaned)
	return strings.TrimSpace(cleaned)
}

func CleanString(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", ErrEmptyField
	}
	return SanitizeText(trimmed), nil
}

func ParsePrice(raw string) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, ErrInvalidPrice
	}

	lower := strings.ToLower(trimmed)
	if lower == "free" || lower == "\"free\"" {
		return 0.0, nil
	}

	cleaned := strings.Trim(trimmed, "$€£¥ \"'")
	cleaned = strings.TrimSpace(cleaned)

	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil || val < 0 {
		return 0, ErrInvalidPrice
	}

	return val, nil
}

func ParseStock(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, "\"'")
	if trimmed == "" {
		return 0, ErrInvalidStock
	}

	val, err := strconv.Atoi(trimmed)
	if err != nil || val < 0 {
		return 0, ErrInvalidStock
	}

	return val, nil
}

func ParseWeight(raw string) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, "\"'")
	if trimmed == "" {
		return 0, ErrInvalidWeight
	}

	val, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || val < 0 {
		return 0, ErrInvalidWeight
	}

	return val, nil
}
