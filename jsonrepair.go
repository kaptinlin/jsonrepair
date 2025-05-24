package jsonrepair

import (
	"fmt"
	"regexp"
	"strings"
)

// JSONRepair attempts to repair the given JSON string and returns the repaired version.
func JSONRepair(text string) (string, error) {
	runes := []rune(text)
	i := 0
	var output strings.Builder

	if !parseValue(&runes, &i, &output) {
		return "", fmt.Errorf("%w at position %d", ErrUnexpectedEnd, len(runes))
	}

	processedComma := parseCharacter(&runes, &i, &output, codeComma)
	if processedComma {
		parseWhitespaceAndSkipComments(&runes, &i, &output)
	}

	if i < len(runes) && isStartOfValue(runes[i]) && endsWithCommaOrNewline(output.String()) {
		if !processedComma {
			outputStr := insertBeforeLastWhitespace(output.String(), ",")
			output.Reset()
			output.WriteString(outputStr)
		}
		parseNewlineDelimitedJSON(&runes, &i, &output)
	} else if processedComma {
		outputStr := stripLastOccurrence(output.String(), ",", false)
		output.Reset()
		output.WriteString(outputStr)
	}

	// repair redundant end quotes
	for i < len(runes) && (runes[i] == codeClosingBrace || runes[i] == codeClosingBracket) {
		i++
		parseWhitespaceAndSkipComments(&runes, &i, &output)
	}

	if i >= len(runes) {
		return output.String(), nil
	}

	return "", fmt.Errorf("%w: '%c' at position %d", ErrUnexpectedCharacter, runes[i], i)
}

// parseValue determines the type of the next value in the input text and parses it accordingly.
func parseValue(text *[]rune, i *int, output *strings.Builder) bool {
	parseWhitespaceAndSkipComments(text, i, output)

	processed := parseObject(text, i, output) ||
		parseArray(text, i, output) ||
		parseString(text, i, output, false) ||
		parseNumber(text, i, output) ||
		parseKeywords(text, i, output) ||
		parseUnquotedString(text, i, output)
	parseWhitespaceAndSkipComments(text, i, output)
	return processed
}

// parseWhitespaceAndSkipComments parses whitespace and skips comments.
func parseWhitespaceAndSkipComments(text *[]rune, i *int, output *strings.Builder) bool {
	start := *i
	parseWhitespace(text, i, output)
	for {
		changed := parseComment(text, i)
		if changed {
			changed = parseWhitespace(text, i, output)
		}

		if !changed {
			break
		}
	}

	return *i > start
}

// parseWhitespace parses whitespace characters.
func parseWhitespace(text *[]rune, i *int, output *strings.Builder) bool {
	start := *i
	whitespace := strings.Builder{}
	for *i < len(*text) && (isWhitespace((*text)[*i]) || isSpecialWhitespace((*text)[*i])) {
		if isWhitespace((*text)[*i]) {
			whitespace.WriteRune((*text)[*i])
		} else {
			whitespace.WriteRune(' ') // repair special whitespace
		}
		*i++
	}
	if whitespace.Len() > 0 {
		output.WriteString(whitespace.String())
		return true
	}
	return *i > start
}

// parseComment parses both single-line (//) and multi-line (/* */) comments.
func parseComment(text *[]rune, i *int) bool {
	if *i+1 < len(*text) {
		if (*text)[*i] == codeSlash && (*text)[*i+1] == codeAsterisk { // multi-line comment
			// repair block comment by skipping it
			for *i < len(*text) && !atEndOfBlockComment(text, i) {
				*i++
			}
			if *i+2 <= len(*text) {
				*i += 2 // move past the end of the block comment
			}
			return true
		} else if (*text)[*i] == codeSlash && (*text)[*i+1] == codeSlash { // single-line comment
			// repair line comment by skipping it
			for *i < len(*text) && (*text)[*i] != codeNewline {
				*i++
			}
			return true
		}
	}
	return false
}

// parseCharacter parses a specific character and adds it to the output if it matches the expected code.
func parseCharacter(text *[]rune, i *int, output *strings.Builder, code rune) bool {
	if *i < len(*text) && (*text)[*i] == code {
		output.WriteRune((*text)[*i])
		*i++
		return true
	}
	return false
}

// skipCharacter skips a specific character in the input text if it matches the expected code.
func skipCharacter(text *[]rune, i *int, code rune) bool {
	if *i < len(*text) && (*text)[*i] == code {
		*i++
		return true
	}
	return false
}

// skipEscapeCharacter skips an escape character in the input text.
func skipEscapeCharacter(text *[]rune, i *int) bool {
	return skipCharacter(text, i, codeBackslash)
}

// skipEllipsis skips ellipsis (three dots) in arrays or objects.
func skipEllipsis(text *[]rune, i *int, output *strings.Builder) bool {
	parseWhitespaceAndSkipComments(text, i, output)

	if *i+2 < len(*text) &&
		(*text)[*i] == codeDot &&
		(*text)[*i+1] == codeDot &&
		(*text)[*i+2] == codeDot {
		*i += 3
		parseWhitespaceAndSkipComments(text, i, output)
		skipCharacter(text, i, codeComma)
		return true
	}
	return false
}

// parseObject parses an object from the input text.
func parseObject(text *[]rune, i *int, output *strings.Builder) bool {
	if *i < len(*text) && (*text)[*i] == codeOpeningBrace {
		output.WriteRune((*text)[*i])
		*i++
		parseWhitespaceAndSkipComments(text, i, output)

		// repair: skip leading comma like in {, message: "hi"}
		if skipCharacter(text, i, codeComma) {
			parseWhitespaceAndSkipComments(text, i, output)
		}

		initial := true
		for *i < len(*text) && (*text)[*i] != codeClosingBrace {
			var processedComma bool
			if !initial {
				processedComma = parseCharacter(text, i, output, codeComma)
				if !processedComma {
					// repair missing comma
					outputStr := insertBeforeLastWhitespace(output.String(), ",")
					output.Reset()
					output.WriteString(outputStr)
				}
				parseWhitespaceAndSkipComments(text, i, output)
			} else {
				processedComma = true
				initial = false
			}

			skipEllipsis(text, i, output)

			processedKey := parseString(text, i, output, false) || parseUnquotedString(text, i, output)
			if !processedKey {
				if *i >= len(*text) ||
					(*text)[*i] == codeClosingBrace ||
					(*text)[*i] == codeOpeningBrace ||
					(*text)[*i] == codeClosingBracket ||
					(*text)[*i] == codeOpeningBracket ||
					(*text)[*i] == 0 {
					// repair trailing comma
					outputStr := stripLastOccurrence(output.String(), ",", false)
					output.Reset()
					output.WriteString(outputStr)
					break
				} else {
					// throwObjectKeyExpected() equivalent
					return false
				}
			}

			parseWhitespaceAndSkipComments(text, i, output)
			processedColon := parseCharacter(text, i, output, codeColon)
			truncatedText := *i >= len(*text)
			if !processedColon {
				if *i < len(*text) && isStartOfValue((*text)[*i]) || truncatedText {
					// repair missing colon
					outputStr := insertBeforeLastWhitespace(output.String(), ":")
					output.Reset()
					output.WriteString(outputStr)
				} else {
					// throwColonExpected() equivalent
					return false
				}
			}

			processedValue := parseValue(text, i, output)
			if !processedValue {
				if processedColon || truncatedText {
					// repair missing object value
					output.WriteString("null")
				} else {
					// throwColonExpected() equivalent
					return false
				}
			}
		}

		if *i < len(*text) && (*text)[*i] == codeClosingBrace {
			output.WriteRune((*text)[*i])
			*i++
		} else {
			// repair missing end bracket
			outputStr := insertBeforeLastWhitespace(output.String(), "}")
			output.Reset()
			output.WriteString(outputStr)
		}
		return true
	}
	return false
}

// parseArray parses an array from the input text.
func parseArray(text *[]rune, i *int, output *strings.Builder) bool {
	if *i >= len(*text) {
		return false
	}

	if (*text)[*i] == codeOpeningBracket {
		output.WriteRune((*text)[*i])
		*i++
		parseWhitespaceAndSkipComments(text, i, output)

		if skipCharacter(text, i, codeComma) {
			parseWhitespaceAndSkipComments(text, i, output)
		}

		initial := true
		for *i < len(*text) && (*text)[*i] != codeClosingBracket {
			if !initial {
				processedComma := parseCharacter(text, i, output, codeComma)
				if !processedComma {
					outputStr := insertBeforeLastWhitespace(output.String(), ",")
					output.Reset()
					output.WriteString(outputStr)
				}
			} else {
				initial = false
			}

			skipEllipsis(text, i, output)

			processedValue := parseValue(text, i, output)

			if !processedValue {
				// repair trailing comma
				outputStr := stripLastOccurrence(output.String(), ",", false)
				output.Reset()
				output.WriteString(outputStr)
				break
			}
		}

		if *i < len(*text) && (*text)[*i] == codeClosingBracket {
			output.WriteRune((*text)[*i])
			*i++
		} else {
			// repair missing closing array bracket
			outputStr := insertBeforeLastWhitespace(output.String(), "]")
			output.Reset()
			output.WriteString(outputStr)
		}
		return true
	}
	return false
}

// parseNewlineDelimitedJSON parses Newline Delimited JSON (NDJSON) from the input text.
func parseNewlineDelimitedJSON(text *[]rune, i *int, output *strings.Builder) {
	initial := true
	processedValue := true

	for processedValue {
		if !initial {
			// parse optional comma, insert when missing
			processedComma := parseCharacter(text, i, output, codeComma)
			if !processedComma {
				// repair: add missing comma
				outputStr := insertBeforeLastWhitespace(output.String(), ",")
				output.Reset()
				output.WriteString(outputStr)
			}
		} else {
			initial = false
		}

		processedValue = parseValue(text, i, output)
	}

	if !processedValue {
		// repair: remove trailing comma
		outputStr := stripLastOccurrence(output.String(), ",", false)
		output.Reset()
		output.WriteString(outputStr)
	}

	// repair: wrap the output inside array brackets
	outputStr := fmt.Sprintf("[\n%s\n]", output.String())
	output.Reset()
	output.WriteString(outputStr)
}

// parseString parses a string from the input text, handling various quote and escape scenarios.
func parseString(text *[]rune, i *int, output *strings.Builder, stopAtDelimiter bool) bool {
	if *i >= len(*text) {
		return false
	}

	skipEscapeChars := (*text)[*i] == codeBackslash
	if skipEscapeChars {
		*i++
	}

	if isQuote((*text)[*i]) {
		var isEndQuote func(rune) bool

		startQuote := (*text)[*i]
		isEndQuote = func(code rune) bool {
			switch startQuote {
			case codeDoubleQuote:
				return isDoubleQuote(code)
			case codeQuote:
				return isSingleQuote(code)
			case codeDoubleQuoteLeft, codeDoubleQuoteRight:
				return isDoubleQuoteLike(code)
			case codeQuoteLeft, codeQuoteRight, codeGraveAccent, codeAcuteAccent:
				return isSingleQuoteLike(code)
			default:
				return code == startQuote
			}
		}

		iBefore := *i
		oBefore := output.Len()

		str := strings.Builder{}
		str.WriteRune('"')
		*i++

		for {
			if *i >= len(*text) {
				// end of text, we are missing an end quote

				iPrev := prevNonWhitespaceIndex(*text, *i-1)
				if !stopAtDelimiter && isDelimiter((*text)[iPrev]) {
					// if the text ends with a delimiter, like ["hello],
					// so the missing end quote should be inserted before this delimiter
					// retry parsing the string, stopping at the first next delimiter
					*i = iBefore
					tempOutput := output.String()[:oBefore]
					output.Reset()
					output.WriteString(tempOutput)
					return parseString(text, i, output, true)
				}

				// repair missing quote
				output.WriteString(insertBeforeLastWhitespace(str.String(), "\""))
				return true
			} else if isEndQuote((*text)[*i]) {
				// end quote
				// let us check what is before and after the quote to verify whether this is a legit end quote
				iQuote := *i
				oQuote := str.Len()
				str.WriteRune('"')
				*i++
				output.WriteString(str.String())

				parseWhitespaceAndSkipComments(text, i, output)

				if stopAtDelimiter || *i >= len(*text) || isDelimiter((*text)[*i]) || isQuote((*text)[*i]) || isDigit((*text)[*i]) {
					// The quote is followed by the end of the text, a delimiter, or a next value
					// so the quote is indeed the end of the string
					parseConcatenatedString(text, i, output)
					return true
				}

				if isDelimiter((*text)[prevNonWhitespaceIndex(*text, iQuote-1)]) {
					// This is not the right end quote: it is preceded by a delimiter,
					// and NOT followed by a delimiter. So, there is an end quote missing
					// parse the string again and then stop at the first next delimiter
					*i = iBefore
					tempOutput := output.String()[:oBefore]
					output.Reset()
					output.WriteString(tempOutput)
					return parseString(text, i, output, true)
				}

				// revert to right after the quote but before any whitespace, and continue parsing the string
				if oBefore <= output.Len() {
					tempOutput := output.String()
					output.Reset()
					output.WriteString(tempOutput[:oBefore])
				}
				*i = iQuote + 1

				// repair unescaped quote
				if oQuote <= str.Len() {
					tempStr := str.String()
					str.Reset()
					str.WriteString(tempStr[:oQuote])
					str.WriteRune('\\')
					str.WriteString(tempStr[oQuote:])
				}
			} else if stopAtDelimiter && isDelimiter((*text)[*i]) {
				// we're in the mode to stop the string at the first delimiter
				// because there is an end quote missing

				// repair missing quote
				output.WriteString(insertBeforeLastWhitespace(str.String(), "\""))
				parseConcatenatedString(text, i, output)
				return true
			} else if (*text)[*i] == codeBackslash {
				// handle escaped content like \n or \u2605
				if *i+1 >= len(*text) {
					return false
				}
				char := (*text)[*i+1]
				_, exists := escapeCharacters[char]
				if exists {
					str.WriteRune('\\') // different from the original code
					str.WriteRune(char)
					*i += 2
				} else if char == 'u' {
					// Handling Unicode escape sequence \uXXXX
					j := 2
					for j < 6 && *i+j < len(*text) && isHex((*text)[*i+j]) {
						j++
					}

					if j == 6 {
						// Valid Unicode escape sequence
						unicodeStr := string((*text)[*i : *i+6])
						str.WriteString(unicodeStr)
						*i += 6
					} else if *i+j >= len(*text) {
						// repair invalid or truncated Unicode char at the end of the text
						// by removing the Unicode char and ending the string here
						*i = len(*text)
					} else {
						// repair invalid Unicode character: remove it
						str.WriteRune('\\')
						str.WriteRune('u')
						*i += 2
					}
				} else {
					str.WriteRune(char)
					*i += 2
				}
			} else {
				// handle regular characters
				char := (*text)[*i]
				code := (*text)[*i]
				if code == codeDoubleQuote && (*text)[*i-1] != codeBackslash {
					// repair unescaped double quote
					str.WriteRune('\\')
					str.WriteRune(char)
					*i++
				} else if isControlCharacter(code) {
					// unescaped control character
					str.WriteString(controlCharacters[code])
					*i++
				} else {
					if !isValidStringCharacter(code) {
						return false // different from the original code
					}
					str.WriteRune(char)
					*i++
				}
			}
			if skipEscapeChars {
				// repair: skipped escape character (nothing to do)
				skipEscapeCharacter(text, i)
			}
		}
	}
	return false
}

// parseConcatenatedString parses and repairs concatenated strings (e.g., "hello" + "world").
func parseConcatenatedString(text *[]rune, i *int, output *strings.Builder) bool {
	processed := false

	parseWhitespaceAndSkipComments(text, i, output)
	for *i < len(*text) && (*text)[*i] == '+' {
		processed = true
		*i++
		parseWhitespaceAndSkipComments(text, i, output)

		// Repair: remove the end quote of the first string
		outputString := output.String()
		lastQuoteIndex := strings.LastIndex(outputString, "\"")
		if lastQuoteIndex != -1 {
			output.Reset()
			output.WriteString(outputString[:lastQuoteIndex])
		}

		start := output.Len()
		if parseString(text, i, output, false) {
			// Repair: remove the start quote of the second string
			outputString = output.String()
			if start < len(outputString) {
				output.Reset()
				output.WriteString(removeAtIndex(outputString, start, 1))
			}
		} else {
			// Repair: remove the + because it is not followed by a string
			output.WriteString("\"")
		}
	}

	return processed
}

// parseNumber parses a number from the input text, handling various numeric formats.
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

	// Note that in JSON leading zeros like "00789" are not allowed.
	// We will allow all leading zeros here though and at the end of parseNumber
	// check against trailing zeros and repair that if needed.
	// Leading zeros can have meaning, so we should not clear them.
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
		hasInvalidLeadingZero := regexp.MustCompile(`^0\d`).MatchString(num)
		if hasInvalidLeadingZero {
			fmt.Fprintf(output, `"%s"`, num)
		} else {
			output.WriteString(num)
		}
		return true
	}
	return false
}

// parseKeywords parses and repairs JSON keywords (true, false, null) and Python keywords (True, False, None).
func parseKeywords(text *[]rune, i *int, output *strings.Builder) bool {
	return parseKeyword(text, i, output, "true", "true") ||
		parseKeyword(text, i, output, "false", "false") ||
		parseKeyword(text, i, output, "null", "null") ||
		parseKeyword(text, i, output, "True", "true") ||
		parseKeyword(text, i, output, "False", "false") ||
		parseKeyword(text, i, output, "None", "null")
}

// parseKeyword parses a specific keyword from the input text.
func parseKeyword(text *[]rune, i *int, output *strings.Builder, name, value string) bool {
	if len(*text)-*i >= len(name) && string((*text)[*i:*i+len(name)]) == name {
		output.WriteString(value)
		*i += len(name)
		return true
	}
	return false
}

// parseUnquotedString parses and repairs unquoted strings, MongoDB function calls, and JSONP function calls.
func parseUnquotedString(text *[]rune, i *int, output *strings.Builder) bool {
	start := *i
	// Move the index forward until a delimiter or quote is found
	for *i < len(*text) && !isDelimiterExceptSlash((*text)[*i]) && !isQuote((*text)[*i]) {
		*i++
	}

	if *i > start {
		// Check for MongoDB function call or JSONP function call
		trimmedSymbol := strings.TrimSpace(string((*text)[start:*i]))
		if *i < len(*text) && (*text)[*i] == codeOpenParenthesis && isFunctionName(trimmedSymbol) {
			*i++
			parseValue(text, i, output)
			if *i < len(*text) && (*text)[*i] == codeCloseParenthesis {
				*i++
				if *i < len(*text) && (*text)[*i] == codeSemicolon {
					*i++
				}
			}
			return true
		} else {
			// Move back to prevent trailing whitespaces in the string
			for *i > start && isWhitespace((*text)[*i-1]) {
				*i--
			}
			symbol := strings.TrimSpace(string((*text)[start:*i]))
			if symbol == "undefined" {
				output.WriteString("null")
			} else {
				// Ensure special quotes are replaced with double quotes
				repairedSymbol := strings.Builder{}
				for _, char := range symbol {
					if isSingleQuoteLike(char) || isDoubleQuoteLike(char) {
						repairedSymbol.WriteRune('"')
					} else {
						repairedSymbol.WriteRune(char)
					}
				}
				fmt.Fprintf(output, `"%s"`, repairedSymbol.String())
			}
			// Skip the end quote if encountered
			if *i < len(*text) && (*text)[*i] == codeDoubleQuote {
				*i++
			}
			return true
		}
	}
	return false
}
