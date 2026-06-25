package jsonrepair

import "strings"

// parseWhitespaceAndSkipComments parses whitespace and removes comments.
func parseWhitespaceAndSkipComments(text *[]rune, i *int, output *strings.Builder, skipNewline bool) bool {
	start := *i
	parseWhitespace(text, i, output, skipNewline)
	for parseComment(text, i) {
		parseWhitespace(text, i, output, skipNewline)
	}
	return *i > start
}

// parseWhitespace parses whitespace characters.
func parseWhitespace(text *[]rune, i *int, output *strings.Builder, skipNewline bool) bool {
	start := *i
	isW := isWhitespace
	if !skipNewline {
		isW = isWhitespaceExceptNewline
	}

	for *i < len(*text) && (isW((*text)[*i]) || isSpecialWhitespace((*text)[*i])) {
		if isSpecialWhitespace((*text)[*i]) {
			output.WriteByte(' ') // repair special whitespace
		} else {
			output.WriteRune((*text)[*i])
		}
		*i++
	}

	return *i > start
}

// parseComment parses both single-line (//) and multi-line (/* */) comments.
func parseComment(text *[]rune, i *int) bool {
	if *i+1 >= len(*text) || (*text)[*i] != codeSlash {
		return false
	}

	switch (*text)[*i+1] {
	case codeAsterisk:
		for *i < len(*text) && (*i+1 >= len(*text) || (*text)[*i] != codeAsterisk || (*text)[*i+1] != codeSlash) {
			*i++
		}
		if *i+2 <= len(*text) {
			*i += 2
		}
		return true
	case codeSlash:
		for *i < len(*text) && (*text)[*i] != codeNewline {
			*i++
		}
		return true
	default:
		return false
	}
}

// resetOutput replaces the entire output buffer with the given string.
func resetOutput(output *strings.Builder, s string) {
	output.Reset()
	output.WriteString(s)
}

// parseCharacter parses a specific character and adds it to output if it matches.
func parseCharacter(text *[]rune, i *int, output *strings.Builder, code rune) bool {
	if *i < len(*text) && (*text)[*i] == code {
		output.WriteRune((*text)[*i])
		*i++
		return true
	}
	return false
}

// skipCharacter skips a specific character if it matches.
func skipCharacter(text *[]rune, i *int, code rune) bool {
	if *i < len(*text) && (*text)[*i] == code {
		*i++
		return true
	}
	return false
}

// skipEllipsis skips ellipsis (three dots) in arrays or objects.
func skipEllipsis(text *[]rune, i *int, output *strings.Builder) bool {
	parseWhitespaceAndSkipComments(text, i, output, true)

	if *i+2 >= len(*text) || (*text)[*i] != codeDot || (*text)[*i+1] != codeDot || (*text)[*i+2] != codeDot {
		return false
	}

	*i += 3
	parseWhitespaceAndSkipComments(text, i, output, true)
	skipCharacter(text, i, codeComma)
	return true
}
