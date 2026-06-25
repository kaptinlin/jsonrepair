package jsonrepair

import (
	"fmt"
	"strings"
)

// parseString parses a string, handling various quote and escape scenarios.
// Returns (success, error) where error is non-nil for non-repairable issues.
func parseString(text *[]rune, i *int, output *strings.Builder, stopAtDelimiter bool, stopAtIndex int) (bool, error) {
	if *i >= len(*text) {
		return false, nil
	}

	skipEscapeChars := (*text)[*i] == codeBackslash
	if skipEscapeChars {
		// repair: remove the first escape character
		*i++
	}

	if *i < len(*text) && isQuote((*text)[*i]) {
		isEndQuote := endQuoteMatcher((*text)[*i])

		iBefore := *i
		oBefore := output.Len()

		mightContainFilePaths := analyzePotentialFilePath(text, *i)

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
					resetOutput(output, output.String()[:oBefore])
					return parseString(text, i, output, true, -1)
				}

				// repair missing quote
				strStr := insertBeforeLastWhitespace(str.String(), "\"")
				output.WriteString(strStr)
				return true, nil
			}

			if stopAtIndex != -1 && *i == stopAtIndex {
				// use the stop index detected in the first iteration, and repair end quote
				strStr := insertBeforeLastWhitespace(str.String(), "\"")
				output.WriteString(strStr)
				return true, nil
			}

			switch {
			case isEndQuote((*text)[*i]):
				// end quote
				iQuote := *i
				oQuote := str.Len()
				str.WriteRune('"')
				*i++
				output.WriteString(str.String())

				iAfterWhitespace := *i
				var tempWhitespace strings.Builder
				parseWhitespaceAndSkipComments(text, &iAfterWhitespace, &tempWhitespace, false)

				if stopAtDelimiter ||
					iAfterWhitespace >= len(*text) ||
					isDelimiter((*text)[iAfterWhitespace]) ||
					isQuote((*text)[iAfterWhitespace]) ||
					isDigit((*text)[iAfterWhitespace]) {
					// The quote is followed by the end of the text, a delimiter,
					// or a next value. So the quote is indeed the end of the string.
					*i = iAfterWhitespace
					output.WriteString(tempWhitespace.String())
					parseConcatenatedString(text, i, output)
					return true, nil
				}

				iPrevChar := prevNonWhitespaceIndex(*text, iQuote-1)
				if iPrevChar != -1 {
					prevChar := (*text)[iPrevChar]
					switch {
					case prevChar == ',':
						*i = iBefore
						resetOutput(output, output.String()[:oBefore])
						return parseString(text, i, output, false, iPrevChar)
					case isDelimiter(prevChar):
						*i = iBefore
						resetOutput(output, output.String()[:oBefore])
						return parseString(text, i, output, true, -1)
					}
				}

				// revert to right after the quote but before any whitespace, and continue parsing the string
				resetOutput(output, output.String()[:oBefore])
				*i = iQuote + 1

				// repair unescaped quote
				revertedStr := str.String()[:oQuote] + "\\\""
				str.Reset()
				str.WriteString(revertedStr)
			case stopAtDelimiter && isUnquotedStringDelimiter((*text)[*i]):
				// we're in the mode to stop the string at the first delimiter
				// because there is an end quote missing
				if *i > 0 && (*text)[*i-1] == ':' &&
					regexURLStart.MatchString(string((*text)[iBefore+1:min(*i+2, len(*text))])) {
					for *i < len(*text) && isURLChar((*text)[*i]) {
						str.WriteRune((*text)[*i])
						*i++
					}
				}

				// repair missing quote
				strStr := insertBeforeLastWhitespace(str.String(), "\"")
				output.WriteString(strStr)
				parseConcatenatedString(text, i, output)
				return true, nil
			case (*text)[*i] == '\\':
				// handle escaped content like \n or \u2605
				if *i+1 >= len(*text) {
					// repair: incomplete escape sequence at end of string
					// just remove the backslash and end the string
					strStr := insertBeforeLastWhitespace(str.String(), "\"")
					output.WriteString(strStr)
					*i++
					return true, nil
				}

				char := (*text)[*i+1]
				if _, ok := escapeCharacters[char]; ok {
					if mightContainFilePaths {
						// In file path context, escape the backslash as literal
						str.WriteString("\\\\")
						*i++
					} else {
						// Valid JSON escape character - keep as is
						str.WriteRune((*text)[*i])
						str.WriteRune((*text)[*i+1])
						*i += 2
					}

					break
				}

				if char == 'u' {
					// Handle Unicode escape sequences
					const unicodeEscapeLen = 6
					const hexDigits = 4
					j := 2
					hexCount := 0
					// Count valid hex characters
					for j < unicodeEscapeLen && *i+j < len(*text) && isHex((*text)[*i+j]) {
						j++
						hexCount++
					}

					switch {
					case hexCount == hexDigits:
						if mightContainFilePaths {
							// In file path context, escape the backslash as literal
							str.WriteString("\\\\")
							*i++
						} else {
							// Valid Unicode escape sequence - keep as is
							str.WriteString(string((*text)[*i : *i+unicodeEscapeLen]))
							*i += unicodeEscapeLen
						}
					case *i+j >= len(*text):
						// repair invalid or truncated unicode char at the end of the text
						// by removing the unicode char and ending the string here
						*i = len(*text)
					default:
						// Invalid Unicode escape sequence
						if mightContainFilePaths && hexCount == 0 && *i+2 < len(*text) {
							// In file path context, \u followed by non-hex might be literal backslash
							// For example: \users, \util, etc.
							nextChar := (*text)[*i+2]
							if (nextChar >= 'a' && nextChar <= 'z') || (nextChar >= 'A' && nextChar <= 'Z') {
								// Looks like \users, \util - treat as literal backslash
								str.WriteString("\\\\")
								*i++
							} else {
								return false, invalidUnicodeSequenceError(*text, *i, hexCount, false)
							}
						} else {
							return false, invalidUnicodeSequenceError(*text, *i, hexCount, true)
						}
					}

					break
				}

				if char == codeNewline {
					str.WriteString(`\n`)
					*i += 2
				} else {
					if stopAtIndex != -1 && *i == stopAtIndex-1 && isDelimiter((*text)[stopAtIndex]) {
						// stop before the delimiter that triggered reparsing to avoid infinite recursion
						output.WriteString(insertBeforeLastWhitespace(str.String(), "\""))
						*i = stopAtIndex
						return true, nil
					}

					if mightContainFilePaths {
						// In file path context, escape the backslash as literal
						str.WriteString("\\\\")
						*i++
					} else {
						// Default behavior: remove invalid escape character
						str.WriteRune(char)
						*i += 2
					}
				}
			default:
				// handle regular characters
				char := (*text)[*i]
				switch {
				case char == '"' && (*text)[*i-1] != '\\':
					// repair unescaped double quote
					str.WriteString("\\\"")
					*i++
				case isControlCharacter(char):
					// unescaped control character
					if replacement, ok := controlCharacters[char]; ok {
						str.WriteString(replacement)
					}
					*i++
				default:
					// Check character validity - matches TypeScript throwInvalidCharacter()
					if !isValidStringCharacter(char) {
						// Format control characters as Unicode escape sequences to match TypeScript
						message := fmt.Sprintf("invalid character \"\\\\u%04x\"", char)
						return false, newInvalidCharacterError(message, *i)
					}
					str.WriteRune(char)
					*i++
				}
			}

			if skipEscapeChars {
				// repair: skipped escape character (nothing to do)
				skipCharacter(text, i, codeBackslash)
			}
		}
	}

	return false, nil
}

func invalidUnicodeSequenceError(text []rune, start, hexCount int, quoteIncomplete bool) *Error {
	const unicodeEscapeLen = 6

	end := 2
	for end < unicodeEscapeLen && start+end < len(text) {
		nextChar := text[start+end]
		if nextChar == '"' || nextChar == '\'' || isWhitespace(nextChar) {
			break
		}
		end++
	}

	chars := string(text[start : start+end])
	if quoteIncomplete && hexCount < 4 && end == 2+hexCount {
		return newInvalidUnicodeError(fmt.Sprintf("invalid unicode character %q\"", chars), start)
	}
	return newInvalidUnicodeError(fmt.Sprintf("invalid unicode character %q", chars), start)
}

// parseConcatenatedString parses concatenated strings (e.g., "hello" + "world").
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
		resetOutput(output, stripLastOccurrence(output.String(), "\"", true))
		start := output.Len()

		stringProcessed, err := parseString(text, i, output, false, -1)
		if err != nil {
			// For concatenated strings, errors are not critical - just stop processing
			stringProcessed = false
		}
		if stringProcessed {
			// repair: remove the start quote of the second string
			outputStr := output.String()
			if len(outputStr) > start {
				resetOutput(output, outputStr[:start]+outputStr[start+1:])
			}
		} else {
			// repair: remove the + because it is not followed by a string
			resetOutput(output, insertBeforeLastWhitespace(output.String(), "\""))
		}
	}

	if !processed {
		// revert parsing whitespace
		*i = iBeforeWhitespace
		resetOutput(output, output.String()[:oBeforeWhitespace])
	}

	return processed
}
