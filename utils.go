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
func insertBeforeLastWhitespace(text, textToInsert string) string {
	index := len(text)
	if index == 0 || !isWhitespace(rune(text[index-1])) {
		return text + textToInsert
	}
	for index > 0 && isWhitespace(rune(text[index-1])) {
		index--
	}
	return text[:index] + textToInsert + text[index:]
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
	return code >= 0x20 && code <= 0x10FFFF
}

// isDelimiter checks if a character is a delimiter.
func isDelimiter(char rune) bool {
	return regexDelimiter.MatchString(string(char))
}

// Regular expression for delimiters.
var regexDelimiter = regexp.MustCompile(`^[,:[\]/{}()\n+]$`)

// isDelimiterExceptSlash checks if a character is a delimiter except for slash.
func isDelimiterExceptSlash(char rune) bool {
	return isDelimiter(char) && char != '/'
}

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
func endsWithCommaOrNewline(text string) bool {
	return regexp.MustCompile(`[,\n][ \t\r]*$`).MatchString(text)
}

// isFunctionName checks if a string is a valid function name.
func isFunctionName(text string) bool {
	return regexp.MustCompile(`^\w+$`).MatchString(text)
}
