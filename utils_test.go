package jsonrepair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertBeforeLastWhitespace(t *testing.T) {
	tests := []struct {
		text         string
		textToInsert string
		expected     string
	}{
		// Basic cases
		{"abc", "123", "abc123"},
		{"abc ", "123", "abc123 "},
		{"abc  ", "123", "abc123  "},
		{"abc \t\n", "123", "abc123 \t\n"},

		// Trailing whitespace cases
		{"abc\n", "123", "abc123\n"},
		{"abc\t", "123", "abc123\t"},
		{"abc\r\n", "123", "abc123\r\n"},
		{"abc \n\t", "123", "abc123 \n\t"},

		// Edge cases
		{"", "123", "123"},
		{" ", "123", "123 "},
		{"\n", "123", "123\n"},
		{"\t", "123", "123\t"},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			result := insertBeforeLastWhitespace(test.text, test.textToInsert)
			assert.Equal(t, test.expected, result)
		})
	}
}
