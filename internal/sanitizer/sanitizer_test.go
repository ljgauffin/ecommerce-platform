package sanitizer_test

import (
	"testing"

	"github.com/lucasgauffin/ecommerce-platform/internal/sanitizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{"standard float", "89.99", 89.99, false},
		{"integer price", "100", 100.0, false},
		{"currency dollar", "$29.99", 29.99, false},
		{"currency euro", "€18.75", 18.75, false},
		{"currency pound with spaces", " £ 50.00 ", 50.00, false},
		{"free lowercase", "free", 0.0, false},
		{"free uppercase", "FREE", 0.0, false},
		{"free quoted", "\"free\"", 0.0, false},
		{"negative price", "-15.00", 0.0, true},
		{"currency negative", "$-5.00", 0.0, true},
		{"empty string", "", 0.0, true},
		{"whitespace only", "   ", 0.0, true},
		{"invalid characters", "N/A", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := sanitizer.ParsePrice(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestParseStock(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{"standard integer", "150", 150, false},
		{"zero stock", "0", 0, false},
		{"stock with quotes", "\"75\"", 75, false},
		{"negative stock", "-10", 0, true},
		{"empty string", "", 0, true},
		{"float string", "12.5", 0, true},
		{"letters", "in-stock", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := sanitizer.ParseStock(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestParseWeight(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{"standard float", "0.35", 0.35, false},
		{"integer weight", "1", 1.0, false},
		{"zero weight", "0", 0.0, false},
		{"weight with quotes", "\"5.20\"", 5.20, false},
		{"negative weight", "-0.5", 0.0, true},
		{"empty string", "", 0.0, true},
		{"whitespace only", "   ", 0.0, true},
		{"letters", "heavy", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := sanitizer.ParseWeight(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Running Shoes",
			expected: "Running Shoes",
		},
		{
			name:     "title with quotes inside",
			input:    `Monitor 27" 4K`,
			expected: `Monitor 27&#34; 4K`,
		},
		{
			name:     "script injection attempt",
			input:    "<script>alert('xss')</script> Cool Gadget",
			expected: "Cool Gadget",
		},
		{
			name:     "html tags stripped",
			input:    "<b>Bold</b> <i>Italic</i> Shoes",
			expected: "Bold Italic Shoes",
		},
		{
			name:     "img onerror injection",
			input:    `<img src=x onerror=alert(1)> Mouse`,
			expected: "Mouse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := sanitizer.SanitizeText(tt.input)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestCleanString(t *testing.T) {
	_, err := sanitizer.CleanString("")
	assert.ErrorIs(t, err, sanitizer.ErrEmptyField)

	_, err = sanitizer.CleanString("   ")
	assert.ErrorIs(t, err, sanitizer.ErrEmptyField)

	cleaned, err := sanitizer.CleanString(" Valid Name ")
	require.NoError(t, err)
	assert.Equal(t, "Valid Name", cleaned)
}
