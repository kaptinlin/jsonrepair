package jsonrepair

import (
	"strings"

	"github.com/go-json-experiment/json"
)

// parseNumber parses a number, handling various numeric formats.
func parseNumber(text *[]rune, i *int, output *strings.Builder) bool {
	start := *i
	if *i < len(*text) && (*text)[*i] == codeMinus {
		*i++
		if atEndOfNumber(text, i) {
			repairNumberEndingWithNumericSymbol(text, start, i, output)
			return true
		}
		if !isDigit((*text)[*i]) {
			*i = start
			return false
		}
	}

	// Preserve leading zeros by quoting the token instead of rewriting its value.
	for *i < len(*text) && isDigit((*text)[*i]) {
		*i++
	}

	if *i < len(*text) && (*text)[*i] == codeDot {
		*i++
		if atEndOfNumber(text, i) {
			repairNumberEndingWithNumericSymbol(text, start, i, output)
			return true
		}
		if !isDigit((*text)[*i]) {
			*i = start
			return false
		}
		for *i < len(*text) && isDigit((*text)[*i]) {
			*i++
		}
	}

	if *i < len(*text) && ((*text)[*i] == codeLowercaseE || (*text)[*i] == codeUppercaseE) {
		*i++
		if *i < len(*text) && ((*text)[*i] == codeMinus || (*text)[*i] == codePlus) {
			*i++
		}
		if atEndOfNumber(text, i) {
			repairNumberEndingWithNumericSymbol(text, start, i, output)
			return true
		}
		if !isDigit((*text)[*i]) {
			*i = start
			return false
		}
		for *i < len(*text) && isDigit((*text)[*i]) {
			*i++
		}
	}

	if !atEndOfNumber(text, i) {
		*i = start
		return false
	}

	if *i > start {
		num := string((*text)[start:*i])
		hasInvalidLeadingZero := leadingZeroRe.MatchString(num)
		if hasInvalidLeadingZero {
			output.WriteByte('"')
			output.WriteString(num)
			output.WriteByte('"')
		} else {
			output.WriteString(num)
		}
		return true
	}
	return false
}

// jsonKeywords maps JSON and Python keyword names to their JSON equivalents.
var jsonKeywords = []struct{ name, value string }{
	{"true", "true"}, {"false", "false"}, {"null", "null"},
	{"True", "true"}, {"False", "false"}, {"None", "null"},
}

// parseKeywords parses JSON keywords (true, false, null) and Python keywords (True, False, None).
func parseKeywords(text *[]rune, i *int, output *strings.Builder) bool {
	for _, kw := range jsonKeywords {
		if parseKeyword(text, i, output, kw.name, kw.value) {
			return true
		}
	}
	return false
}

// parseKeyword parses a specific keyword from the input text.
func parseKeyword(text *[]rune, i *int, output *strings.Builder, name, value string) bool {
	end := *i + len(name)
	if end > len(*text) || string((*text)[*i:end]) != name {
		return false
	}
	if end < len(*text) && !isDelimiter((*text)[end]) && !isWhitespace((*text)[end]) && !isSpecialWhitespace((*text)[end]) {
		return false
	}

	output.WriteString(value)
	*i = end
	return true
}

func hasURLPrefix(text []rune, start int) bool {
	for _, prefix := range []string{"https://", "http://", "ftp://"} {
		end := start + len(prefix)
		if end <= len(text) && strings.EqualFold(string(text[start:end]), prefix) {
			return true
		}
	}
	return false
}

// parseUnquotedStringWithMode parses unquoted strings with a mode parameter to control URL parsing.
func parseUnquotedStringWithMode(text *[]rune, i *int, output *strings.Builder, isKey bool) (bool, error) {
	start := *i

	if *i >= len(*text) {
		return false, nil
	}

	// Check for function name start (MongoDB/JSONP function calls)
	if isFunctionNameCharStart((*text)[*i]) {
		for *i < len(*text) && isFunctionNameChar((*text)[*i]) {
			*i++
		}

		j := *i
		for j < len(*text) && (isWhitespace((*text)[j]) || isSpecialWhitespace((*text)[j])) {
			j++
		}

		if j < len(*text) && (*text)[j] == codeOpenParenthesis {
			*i = j + 1
			if _, err := parseValue(text, i, output); err != nil {
				return false, err
			}

			if *i < len(*text) && (*text)[*i] == codeCloseParenthesis {
				*i++
				if *i < len(*text) && (*text)[*i] == codeSemicolon {
					*i++
				}
			}
			return true, nil
		}
	}

	// Check if this starts with a URL pattern (only when not parsing a key)
	isURL := !isKey && hasURLPrefix(*text, start)

	if isURL {
		for *i < len(*text) && isURLChar((*text)[*i]) {
			*i++
		}
	} else {
		for *i < len(*text) && !isUnquotedStringDelimiter((*text)[*i]) && !isQuote((*text)[*i]) {
			if isKey && (*text)[*i] == codeColon {
				break
			}
			*i++
		}
	}

	if *i <= start {
		return false, nil
	}

	for *i > start && isWhitespace((*text)[*i-1]) {
		*i--
	}

	symbol := string((*text)[start:*i])

	if symbol == "undefined" {
		output.WriteString("null")
	} else {
		output.WriteByte('"')
		for _, char := range symbol {
			if isSingleQuoteLike(char) || isDoubleQuoteLike(char) {
				output.WriteByte('"')
			} else {
				output.WriteRune(char)
			}
		}
		output.WriteByte('"')
	}

	if *i < len(*text) && (*text)[*i] == codeDoubleQuote {
		*i++
	}

	return true, nil
}

func isRegexFlag(char rune) bool {
	switch char {
	case 'd', 'g', 'i', 'm', 's', 'u', 'v', 'y':
		return true
	default:
		return false
	}
}

// parseRegex parses a regex literal like /pattern/flags and wraps it in quotes.
func parseRegex(text *[]rune, i *int, output *strings.Builder) bool {
	if *i >= len(*text) || (*text)[*i] != codeSlash {
		return false
	}

	start := *i
	*i++
	inCharClass := false
	escaped := false
	closed := false

	for *i < len(*text) && !closed {
		char := (*text)[*i]
		switch {
		case escaped:
			escaped = false
		case char == codeBackslash:
			escaped = true
		case char == codeOpeningBracket:
			inCharClass = true
		case char == codeClosingBracket:
			inCharClass = false
		case char == codeSlash && !inCharClass:
			closed = true
		}
		*i++
	}

	for *i < len(*text) && isRegexFlag((*text)[*i]) {
		*i++
	}

	// json.Marshal properly escapes quotes, backslashes, and other special
	// characters in the regex content, preventing XSS when repaired JSON is
	// parsed with eval. See josdejong/jsonrepair#150.
	// json.Marshal cannot fail for a valid UTF-8 string derived from runes.
	regexContent := string((*text)[start:*i])
	jsonBytes, _ := json.Marshal(regexContent)
	output.Write(jsonBytes)
	return true
}
