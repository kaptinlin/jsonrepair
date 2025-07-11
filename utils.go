package jsonrepair

import (
	"regexp"
	"strings"
)

// prevNonWhitespaceIndex finds the previous non-whitespace index in the string.
func prevNonWhitespaceIndex(text []rune, startIndex int) int {
	prev := startIndex
	for prev >= 0 && isWhitespace(text[prev]) {
		prev--
	}
	return prev
}

// atEndOfBlockComment checks if the current position is at the end of a block comment.
func atEndOfBlockComment(text *[]rune, i *int) bool {
	return *i+1 < len(*text) && (*text)[*i] == codeAsterisk && (*text)[*i+1] == codeSlash
}

// atEndOfNumber checks if the end of a number has been reached in the input text.
func atEndOfNumber(text *[]rune, i *int) bool {
	return *i >= len(*text) || isDelimiter((*text)[*i]) || isWhitespace((*text)[*i])
}

// repairNumberEndingWithNumericSymbol repairs numbers cut off at the end.
func repairNumberEndingWithNumericSymbol(text *[]rune, start int, i *int, output *strings.Builder) {
	output.WriteString(string((*text)[start:*i]) + "0")
}

// stripLastOccurrence removes the last occurrence of a specific substring from the input text.
func stripLastOccurrence(text, textToStrip string, stripRemainingText bool) string {
	index := strings.LastIndex(text, textToStrip)
	if index != -1 {
		if stripRemainingText {
			return text[:index]
		}
		return text[:index] + text[index+len(textToStrip):]
	}
	return text
}

// insertBeforeLastWhitespace inserts a substring before the last whitespace in the input text.
// For comma insertion, we want to insert after the value but before any trailing whitespace.
func insertBeforeLastWhitespace(s, textToInsert string) string {
	// If the last character is not whitespace, simply append the text to insert.
	if len(s) == 0 || !isWhitespace(rune(s[len(s)-1])) {
		return s + textToInsert
	}

	// Walk backwards over all trailing whitespace characters (space, tab, cr, lf).
	index := len(s) - 1
	for index >= 0 {
		if !isWhitespace(rune(s[index])) {
			break
		}
		index--
	}

	// index now points at the last non-whitespace character.
	return s[:index+1] + textToInsert + s[index+1:]
}

// removeAtIndex removes a substring from the input text at a specific index.
func removeAtIndex(text string, start, count int) string {
	return text[:start] + text[start+count:]
}

// isHex checks if a rune is a hexadecimal digit.
func isHex(code rune) bool {
	return (code >= codeZero && code <= codeNine) ||
		(code >= codeUppercaseA && code <= codeUppercaseF) ||
		(code >= codeLowercaseA && code <= codeLowercaseF)
}

// isDigit checks if a rune is a digit.
func isDigit(code rune) bool {
	return code >= codeZero && code <= codeNine
}

// isValidStringCharacter checks if a code is a valid string character.
func isValidStringCharacter(code rune) bool {
	// based on https://www.rfc-editor.org/rfc/rfc8259.html#section-7
	return code >= 0x0020 && code <= 0x10FFFF
}

// isDelimiter checks if a character is a delimiter.
func isDelimiter(char rune) bool {
	return regexDelimiter.MatchString(string(char))
}

// regexDelimiter matches a single JSON delimiter character used to separate tokens.
// The character class explicitly lists all delimiter characters and escapes special
// characters to prevent unintended character ranges (e.g. ":[" would otherwise
// create a range from ':' to '[').
var regexDelimiter = regexp.MustCompile(`^[,:\[\]/{}()\n\+]$`)

// isStartOfValue checks if a rune is the start of a JSON value.
func isStartOfValue(char rune) bool {
	return regexStartOfValue.MatchString(string(char)) || isQuote(char)
}

// regexStartOfValue defines the regular expression for the start of a JSON value.
var regexStartOfValue = regexp.MustCompile(`^[{[\w-]$`)

// isControlCharacter checks if a rune is a control character.
func isControlCharacter(code rune) bool {
	return code == codeNewline ||
		code == codeReturn ||
		code == codeTab ||
		code == codeBackspace ||
		code == codeFormFeed
}

// isWhitespace checks if a rune is a whitespace character.
func isWhitespace(code rune) bool {
	return code == codeSpace ||
		code == codeNewline ||
		code == codeTab ||
		code == codeReturn
}

// isSpecialWhitespace checks if a rune is a special whitespace character.
func isSpecialWhitespace(code rune) bool {
	return code == codeNonBreakingSpace ||
		(code >= codeEnQuad && code <= codeHairSpace) ||
		code == codeNarrowNoBreakSpace ||
		code == codeMediumMathematicalSpace ||
		code == codeIdeographicSpace
}

// isQuote checks if a rune is a quote character.
func isQuote(code rune) bool {
	return isDoubleQuoteLike(code) || isSingleQuoteLike(code)
}

// isDoubleQuoteLike checks if a rune is a double quote or a variant of double quote.
func isDoubleQuoteLike(code rune) bool {
	return code == codeDoubleQuote ||
		code == codeDoubleQuoteLeft ||
		code == codeDoubleQuoteRight
}

// isDoubleQuote checks if a rune is a double quote.
func isDoubleQuote(code rune) bool {
	return code == codeDoubleQuote
}

// isSingleQuoteLike checks if a rune is a single quote or a variant of single quote.
func isSingleQuoteLike(code rune) bool {
	return code == codeQuote ||
		code == codeQuoteLeft ||
		code == codeQuoteRight ||
		code == codeGraveAccent ||
		code == codeAcuteAccent
}

// isSingleQuote checks if a rune is a single quote.
func isSingleQuote(code rune) bool {
	return code == codeQuote
}

// endsWithCommaOrNewline checks if the string ends with a comma or newline character and optional whitespace.
// This function should only match commas that are outside of quoted strings.
func endsWithCommaOrNewline(text string) bool {
	if len(text) == 0 {
		return false
	}

	// Find the last non-whitespace character
	runes := []rune(text)
	i := len(runes) - 1

	// Skip trailing whitespace
	for i >= 0 && (runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\r') {
		i--
	}

	if i < 0 {
		return false
	}

	// Check if the last non-whitespace character is a comma or newline
	// But only if it's not inside a quoted string
	if runes[i] == ',' || runes[i] == '\n' {
		// Simple check: if the text ends with a quoted string, the comma is likely inside the string
		// A more robust approach would be to parse the JSON structure, but for now we use a heuristic
		trimmed := strings.TrimSpace(text)
		if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '"' {
			// The text ends with a quote, so any comma before it is likely a JSON separator
			// Look for the pattern: "..." , or "...",
			return regexp.MustCompile(`"[ \t\r]*[,\n][ \t\r]*$`).MatchString(text)
		}
		return true
	}

	return false
}

// isFunctionNameCharStart checks if a rune is a valid function name start character.
func isFunctionNameCharStart(code rune) bool {
	return (code >= 'a' && code <= 'z') || (code >= 'A' && code <= 'Z') || code == '_' || code == '$'
}

// isFunctionNameChar checks if a rune is a valid function name character.
func isFunctionNameChar(code rune) bool {
	return isFunctionNameCharStart(code) || isDigit(code)
}

// isUnquotedStringDelimiter checks if a character is a delimiter for unquoted strings.
func isUnquotedStringDelimiter(char rune) bool {
	return regexUnquotedStringDelimiter.MatchString(string(char))
}

// Similar to regexDelimiter but without ':' since a colon is allowed inside an
// unquoted value until we detect a key/value separator.
var regexUnquotedStringDelimiter = regexp.MustCompile(`^[,\[\]/{}\n\+]$`)

// isWhitespaceExceptNewline checks if a rune is a whitespace character except newline.
func isWhitespaceExceptNewline(code rune) bool {
	return code == codeSpace || code == codeTab || code == codeReturn
}

// URL-related regular expressions and functions
var regexUrlStart = regexp.MustCompile(`^(https?|ftp|mailto|file|data|irc)://`)
var regexUrlChar = regexp.MustCompile(`^[A-Za-z0-9\-._~:/?#@!$&'()*+;=]$`)

// isUrlChar checks if a rune is a valid URL character.
func isUrlChar(code rune) bool {
	return regexUrlChar.MatchString(string(code))
}
