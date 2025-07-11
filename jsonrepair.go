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

	// Parse leading Markdown code block
	parseMarkdownCodeBlock(&runes, &i, []string{"```", "[```", "{```"}, &output)

	if !parseValue(&runes, &i, &output) {
		return "", fmt.Errorf("%w at position %d", ErrUnexpectedEnd, len(runes))
	}

	// Parse trailing Markdown code block
	parseMarkdownCodeBlock(&runes, &i, []string{"```", "```]", "```}"}, &output)

	processedComma := parseCharacter(&runes, &i, &output, codeComma)
	if processedComma {
		parseWhitespaceAndSkipComments(&runes, &i, &output, true)
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
		parseWhitespaceAndSkipComments(&runes, &i, &output, true)
	}

	if i >= len(runes) {
		return output.String(), nil
	}

	return "", fmt.Errorf("%w: '%c' at position %d", ErrUnexpectedCharacter, runes[i], i)
}

// parseValue determines the type of the next value in the input text and parses it accordingly.
func parseValue(text *[]rune, i *int, output *strings.Builder) bool {
	parseWhitespaceAndSkipComments(text, i, output, true)

	processed := parseObject(text, i, output) ||
		parseArray(text, i, output) ||
		parseString(text, i, output, false, -1) ||
		parseNumber(text, i, output) ||
		parseKeywords(text, i, output) ||
		parseUnquotedString(text, i, output) ||
		parseRegex(text, i, output)
	parseWhitespaceAndSkipComments(text, i, output, true)
	return processed
}

// parseWhitespaceAndSkipComments parses whitespace and skips comments.
func parseWhitespaceAndSkipComments(text *[]rune, i *int, output *strings.Builder, skipNewline bool) bool {
	start := *i
	parseWhitespace(text, i, output, skipNewline)
	for {
		changed := parseComment(text, i)
		if changed {
			changed = parseWhitespace(text, i, output, skipNewline)
		}

		if !changed {
			break
		}
	}

	return *i > start
}

// parseWhitespace parses whitespace characters.
func parseWhitespace(text *[]rune, i *int, output *strings.Builder, skipNewline bool) bool {
	start := *i
	whitespace := strings.Builder{}

	isW := isWhitespace
	if !skipNewline {
		isW = isWhitespaceExceptNewline
	}

	for *i < len(*text) && (isW((*text)[*i]) || isSpecialWhitespace((*text)[*i])) {
		if !isSpecialWhitespace((*text)[*i]) {
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
	parseWhitespaceAndSkipComments(text, i, output, true)

	if *i+2 < len(*text) &&
		(*text)[*i] == codeDot &&
		(*text)[*i+1] == codeDot &&
		(*text)[*i+2] == codeDot {
		*i += 3
		parseWhitespaceAndSkipComments(text, i, output, true)
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
		parseWhitespaceAndSkipComments(text, i, output, true)

		// repair: skip leading comma like in {, message: "hi"}
		if skipCharacter(text, i, codeComma) {
			parseWhitespaceAndSkipComments(text, i, output, true)
		}

		initial := true
		for *i < len(*text) && (*text)[*i] != codeClosingBrace {
			if !initial {
				iBefore := *i
				oBefore := output.Len()
				// parse optional comma
				processedComma := parseCharacter(text, i, output, codeComma)
				if processedComma {
					// We just appended the comma, but it may be located *after* a
					// previously written whitespace sequence (for example a
					// newline and indentation). In order to keep the output
					// consistent with the reference implementation, we move the
					// comma so that it comes *before* those trailing
					// whitespaces.
					temp := output.String()
					// Remove the comma we just wrote (it is guaranteed to be
					// the last rune).
					if strings.HasSuffix(temp, ",") {
						temp = temp[:len(temp)-1]
						// Re-insert the comma before the trailing whitespace
						temp = insertBeforeLastWhitespace(temp, ",")

						// After moving the comma, remove the spaces that are
						// still attached to the newline – they will be
						// re-added when we later write the original
						// whitespace found in the source text. This prevents
						// duplicating the indentation (which previously
						// resulted in 4 spaces instead of 2).
						if idx := strings.LastIndex(temp, "\n"); idx != -1 {
							// Only trim spaces when they are *trailing* after the newline.
							j := idx + 1
							for j < len(temp) && (temp[j] == ' ' || temp[j] == '\t') {
								j++
							}
							if j == len(temp) {
								// All remaining characters are whitespace → safe to trim.
								temp = temp[:idx+1]
							}
						}
						output.Reset()
						output.WriteString(temp)
					}
				} else {
					// repair missing comma (original logic)
					*i = iBefore
					tempStr := output.String()
					output.Reset()
					output.WriteString(tempStr[:oBefore])

					outputStr := insertBeforeLastWhitespace(output.String(), ",")
					output.Reset()
					output.WriteString(outputStr)
				}
			} else {
				initial = false
			}

			skipEllipsis(text, i, output)

			processedKey := parseString(text, i, output, false, -1) || parseUnquotedStringWithMode(text, i, output, true)
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
				} else {
					// throwObjectKeyExpected() equivalent
					return false
				}
				break
			}

			parseWhitespaceAndSkipComments(text, i, output, true)
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
		parseWhitespaceAndSkipComments(text, i, output, true)

		if skipCharacter(text, i, codeComma) {
			parseWhitespaceAndSkipComments(text, i, output, true)
		}

		initial := true
		for *i < len(*text) && (*text)[*i] != codeClosingBracket {
			if !initial {
				iBefore := *i
				oBefore := output.Len()
				parseWhitespaceAndSkipComments(text, i, output, true)

				processedComma := parseCharacter(text, i, output, codeComma)
				if !processedComma {
					*i = iBefore
					tempStr := output.String()
					output.Reset()
					output.WriteString(tempStr[:oBefore])

					// repair missing comma
					outputStr := insertBeforeLastWhitespace(output.String(), ",")
					output.Reset()
					output.WriteString(outputStr)
				}
			} else {
				initial = false
			}

			skipEllipsis(text, i, output)

			processedValue := parseValue(text, i, output)

			// Clean up a trailing comma that is **inside** a JSON string when
			// it is directly followed by the string's closing quote. This
			// situation typically comes from an input like "hello,world,"2
			// where the comma actually belongs between two array items but
			// ended up inside the first string. We must *not* touch a string
			// that is literally just a comma (",") – that is a valid value
			// in a JSON array.
			if processedValue {
				outputStr := output.String()

				// We look for ...",\"  (comma just before the closing quote).
				if strings.HasSuffix(outputStr, ",\"") {
					// Ensure the string contains more than just that comma.
					// The minimal string we do NOT want to alter is ",",
					// which would look like ["\",\"]. That has length 3
					// including the comma and quotes -> 4 characters in the
					// output (opening [, closing ], quotes). A safer check is
					// to verify that inside the quotes we have more than one
					// character.

					// Find the position of the opening quote for this value.
					lastQuote := strings.LastIndex(outputStr[:len(outputStr)-2], "\"")
					if lastQuote != -1 && len(outputStr)-2-lastQuote > 2 {
						cleanedStr := outputStr[:len(outputStr)-2] + "\""
						output.Reset()
						output.WriteString(cleanedStr)
					}
				}
			}

			// Note: the TypeScript reference implementation does not attempt to
			// strip trailing commas that are *inside* JSON strings here. Any
			// such cleanup is handled during string parsing itself. Keeping the
			// Go implementation aligned with the reference prevents accidental
			// removal of valid characters such as a standalone "," string.

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
func parseString(text *[]rune, i *int, output *strings.Builder, stopAtDelimiter bool, stopAtIndex int) bool {
	if *i >= len(*text) {
		return false
	}

	skipEscapeChars := (*text)[*i] == codeBackslash
	if skipEscapeChars {
		// repair: remove the first escape character
		*i++
	}

	if *i < len(*text) && isQuote((*text)[*i]) {
		isEndQuote := func(r rune) bool { return r == (*text)[*i] }
		if isDoubleQuote((*text)[*i]) {
			isEndQuote = isDoubleQuote
		} else if isSingleQuote((*text)[*i]) {
			isEndQuote = isSingleQuote
		} else if isSingleQuoteLike((*text)[*i]) {
			isEndQuote = isSingleQuoteLike
		} else if isDoubleQuoteLike((*text)[*i]) {
			isEndQuote = isDoubleQuoteLike
		}

		iBefore := *i
		oBefore := output.Len()

		var str strings.Builder
		str.WriteRune('"')
		*i++

		for {
			if *i >= len(*text) {
				// end of text, we are missing an end quote
				iPrev := prevNonWhitespaceIndex(*text, *i-1)
				if !stopAtDelimiter && iPrev != -1 && isDelimiter((*text)[iPrev]) {
					// if the text ends with a delimiter, like ["hello],
					// so the missing end quote should be inserted before this delimiter
					// retry parsing the string, stopping at the first next delimiter
					*i = iBefore
					tempStr := output.String()
					output.Reset()
					output.WriteString(tempStr[:oBefore])
					return parseString(text, i, output, true, -1)
				}

				// repair missing quote
				strStr := insertBeforeLastWhitespace(str.String(), "\"")
				output.WriteString(strStr)
				return true
			}

			if stopAtIndex != -1 && *i == stopAtIndex {
				// use the stop index detected in the first iteration, and repair end quote
				strStr := insertBeforeLastWhitespace(str.String(), "\"")
				output.WriteString(strStr)
				return true
			}

			if isEndQuote((*text)[*i]) {
				// end quote
				iQuote := *i
				oQuote := str.Len()
				str.WriteRune('"')
				*i++
				output.WriteString(str.String())

				iAfterWhitespace := *i
				var tempWhitespace strings.Builder
				parseWhitespaceAndSkipComments(text, &iAfterWhitespace, &tempWhitespace, false)

				if stopAtDelimiter || iAfterWhitespace >= len(*text) || isDelimiter((*text)[iAfterWhitespace]) || isQuote((*text)[iAfterWhitespace]) || isDigit((*text)[iAfterWhitespace]) {
					// The quote is followed by the end of the text, a delimiter,
					// or a next value. So the quote is indeed the end of the string.
					*i = iAfterWhitespace
					output.WriteString(tempWhitespace.String())
					parseConcatenatedString(text, i, output)
					return true
				}

				iPrevChar := prevNonWhitespaceIndex(*text, iQuote-1)
				if iPrevChar != -1 {
					prevChar := (*text)[iPrevChar]
					if prevChar == ',' {
						*i = iBefore
						tempStr := output.String()
						output.Reset()
						output.WriteString(tempStr[:oBefore])
						return parseString(text, i, output, false, iPrevChar)
					}

					if isDelimiter(prevChar) {
						*i = iBefore
						tempStr := output.String()
						output.Reset()
						output.WriteString(tempStr[:oBefore])
						return parseString(text, i, output, true, -1)
					}
				}

				// revert to right after the quote but before any whitespace, and continue parsing the string
				tempStr := output.String()
				output.Reset()
				output.WriteString(tempStr[:oBefore])
				*i = iQuote + 1

				// repair unescaped quote
				revertedStr := str.String()[:oQuote] + "\\\""
				str.Reset()
				str.WriteString(revertedStr)
			} else if stopAtDelimiter && isUnquotedStringDelimiter((*text)[*i]) {
				// we're in the mode to stop the string at the first delimiter
				// because there is an end quote missing
				if *i > 0 && (*text)[*i-1] == ':' && regexUrlStart.MatchString(string((*text)[iBefore+1:*i+2])) {
					for *i < len(*text) && regexUrlChar.MatchString(string((*text)[*i])) {
						str.WriteRune((*text)[*i])
						*i++
					}
				}

				// repair missing quote
				strStr := insertBeforeLastWhitespace(str.String(), "\"")
				output.WriteString(strStr)
				parseConcatenatedString(text, i, output)
				return true
			} else if (*text)[*i] == '\\' {
				// handle escaped content like \n or \u2605
				char := (*text)[*i+1]
				if _, ok := escapeCharacters[char]; ok {
					str.WriteRune((*text)[*i])
					str.WriteRune((*text)[*i+1])
					*i += 2
				} else if char == 'u' {
					j := 2
					for j < 6 && *i+j < len(*text) && isHex((*text)[*i+j]) {
						j++
					}
					if j == 6 {
						str.WriteString(string((*text)[*i : *i+6]))
						*i += 6
					} else if *i+j >= len(*text) {
						// repair invalid or truncated unicode char at the end of the text
						// by removing the unicode char and ending the string here
						*i = len(*text)
					} else {
						// repair invalid unicode character: remove it
						*i += 2
					}
				} else {
					// repair invalid escape character: remove it
					str.WriteRune(char)
					*i += 2
				}
			} else {
				// handle regular characters
				char := (*text)[*i]
				if char == '"' && (*text)[*i-1] != '\\' {
					// repair unescaped double quote
					str.WriteString("\\\"")
					*i++
				} else if isControlCharacter(char) {
					// unescaped control character
					if replacement, ok := controlCharacters[char]; ok {
						str.WriteString(replacement)
					}
					*i++
				} else {
					if isValidStringCharacter(char) {
						str.WriteRune(char)
					} else {
						// Skip invalid characters or replace with space
						str.WriteRune(' ')
					}
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

	iBeforeWhitespace := *i
	oBeforeWhitespace := output.Len()
	parseWhitespaceAndSkipComments(text, i, output, true)

	for *i < len(*text) && (*text)[*i] == '+' {
		processed = true
		*i++
		parseWhitespaceAndSkipComments(text, i, output, true)

		// repair: remove the end quote of the first string
		outputStr := stripLastOccurrence(output.String(), "\"", true)
		output.Reset()
		output.WriteString(outputStr)
		start := output.Len()

		if parseString(text, i, output, false, -1) {
			// repair: remove the start quote of the second string
			outputStr = output.String()
			if len(outputStr) > start {
				output.Reset()
				output.WriteString(removeAtIndex(outputStr, start, 1))
			}
		} else {
			// repair: remove the + because it is not followed by a string
			outputStr = insertBeforeLastWhitespace(output.String(), "\"")
			output.Reset()
			output.WriteString(outputStr)
		}
	}

	if !processed {
		// revert parsing whitespace
		*i = iBeforeWhitespace
		tempStr := output.String()
		output.Reset()
		output.WriteString(tempStr[:oBeforeWhitespace])
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
	return parseUnquotedStringWithMode(text, i, output, false)
}

// parseUnquotedStringWithMode parses unquoted strings with a mode parameter to control URL parsing
func parseUnquotedStringWithMode(text *[]rune, i *int, output *strings.Builder, isKey bool) bool {
	start := *i

	if *i >= len(*text) {
		return false
	}

	// Check for function name start (MongoDB/JSONP function calls)
	if isFunctionNameCharStart((*text)[*i]) {
		for *i < len(*text) && isFunctionNameChar((*text)[*i]) {
			*i++
		}

		j := *i
		for j < len(*text) && isWhitespace((*text)[j]) {
			j++
		}

		if j < len(*text) && (*text)[j] == codeOpenParenthesis {
			// repair a MongoDB function call like NumberLong("2")
			// repair a JSONP function call like callback({...});
			*i = j + 1

			parseValue(text, i, output)

			if *i < len(*text) && (*text)[*i] == codeCloseParenthesis {
				// repair: skip close bracket of function call
				*i++
				if *i < len(*text) && (*text)[*i] == codeSemicolon {
					// repair: skip semicolon after JSONP call
					*i++
				}
			}

			return true
		}
	}

	// Check if this starts with a URL pattern (only when not parsing a key)
	isURL := false
	if !isKey {
		if start+8 <= len(*text) && string((*text)[start:start+8]) == "https://" {
			isURL = true
		} else if start+7 <= len(*text) && string((*text)[start:start+7]) == "http://" {
			isURL = true
		} else if start+6 <= len(*text) && string((*text)[start:start+6]) == "ftp://" {
			isURL = true
		}
	}

	if isURL {
		// Parse as URL - continue until we hit a proper delimiter (not slash)
		for *i < len(*text) && isUrlChar((*text)[*i]) {
			*i++
		}
	} else {
		// Move the index forward until a delimiter or quote is found
		for *i < len(*text) && !isUnquotedStringDelimiter((*text)[*i]) && !isQuote((*text)[*i]) {
			// If we're parsing a key and encounter a colon, stop here
			if isKey && (*text)[*i] == codeColon {
				break
			}
			*i++
		}
	}

	if *i > start {
		// repair unquoted string
		// also, repair undefined into null

		// first, go back to prevent getting trailing whitespaces in the string
		for *i > start && isWhitespace((*text)[*i-1]) {
			*i--
		}

		symbol := string((*text)[start:*i])

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
	return false
}

// parseRegex parses a regular expression literal like /pattern/flags.
func parseRegex(text *[]rune, i *int, output *strings.Builder) bool {
	if *i < len(*text) && (*text)[*i] == codeSlash {
		start := *i
		*i++

		for *i < len(*text) && ((*text)[*i] != codeSlash || (*text)[*i-1] == codeBackslash) {
			*i++
		}

		if *i < len(*text) && (*text)[*i] == codeSlash {
			*i++
		}

		// Process the regex content to handle escape characters properly
		regexContent := string((*text)[start:*i])
		// Ensure backslashes are properly escaped in the output JSON string
		regexContent = strings.ReplaceAll(regexContent, "\\", "\\\\")

		fmt.Fprintf(output, `"%s"`, regexContent)
		return true
	}
	return false
}

// parseMarkdownCodeBlock parses and skips Markdown fenced code blocks like ``` or ```json
func parseMarkdownCodeBlock(text *[]rune, i *int, blocks []string, output *strings.Builder) bool {
	if skipMarkdownCodeBlock(text, i, blocks) {
		if *i < len(*text) && isFunctionNameCharStart((*text)[*i]) {
			// Strip the optional language specifier like "json"
			for *i < len(*text) && isFunctionNameChar((*text)[*i]) {
				*i++
			}
		}

		// Add any whitespace after code block marker to output
		for *i < len(*text) && (isWhitespace((*text)[*i]) || isSpecialWhitespace((*text)[*i])) {
			if isWhitespace((*text)[*i]) {
				output.WriteRune((*text)[*i])
			} else {
				output.WriteRune(' ') // repair special whitespace
			}
			*i++
		}

		return true
	}
	return false
}

// skipMarkdownCodeBlock checks if we're at a Markdown code block marker and skips it
func skipMarkdownCodeBlock(text *[]rune, i *int, blocks []string) bool {
	for _, block := range blocks {
		blockRunes := []rune(block)
		end := *i + len(blockRunes)
		if end <= len(*text) {
			match := true
			for j := 0; j < len(blockRunes); j++ {
				if (*text)[*i+j] != blockRunes[j] {
					match = false
					break
				}
			}
			if match {
				*i = end
				return true
			}
		}
	}
	return false
}
