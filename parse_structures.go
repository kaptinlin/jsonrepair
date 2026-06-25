package jsonrepair

import "strings"

// parseObject parses a JSON object.
// Returns (success, error) where error is non-nil for non-repairable issues.
func parseObject(text *[]rune, i *int, output *strings.Builder) (bool, error) {
	if *i < len(*text) && (*text)[*i] == codeOpeningBrace {
		output.WriteRune((*text)[*i])
		*i++
		parseWhitespaceAndSkipComments(text, i, output, true)

		// Repair: skip leading comma like in {, message: "hi"}
		skipCharacter(text, i, codeComma)
		parseWhitespaceAndSkipComments(text, i, output, true)

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
					if tempWithoutComma, ok := strings.CutSuffix(temp, ","); ok {
						temp = tempWithoutComma
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
						resetOutput(output, temp)
					}
				} else {
					// repair missing comma - restore output to oBefore and insert comma
					*i = iBefore
					resetOutput(output, insertBeforeLastWhitespace(output.String()[:oBefore], ","))
				}
			} else {
				initial = false
			}

			skipEllipsis(text, i, output)

			stringProcessed, err := parseString(text, i, output, false, -1)
			if err != nil {
				return false, err
			}
			processedKey := stringProcessed
			if !processedKey {
				processedKey, err = parseUnquotedStringWithMode(text, i, output, true)
				if err != nil {
					return false, err
				}
			}
			if !processedKey {
				if *i >= len(*text) ||
					(*text)[*i] == codeClosingBrace ||
					(*text)[*i] == codeOpeningBrace ||
					(*text)[*i] == codeClosingBracket ||
					(*text)[*i] == codeOpeningBracket ||
					(*text)[*i] == 0 {
					// repair trailing comma
					resetOutput(output, stripLastOccurrence(output.String(), ",", false))
				} else {
					return false, newObjectKeyExpectedError(*i)
				}
				break
			}

			parseWhitespaceAndSkipComments(text, i, output, true)
			processedColon := parseCharacter(text, i, output, codeColon)
			truncatedText := *i >= len(*text)
			if !processedColon {
				if (*i < len(*text) && isStartOfValue((*text)[*i])) || truncatedText {
					// repair missing colon
					resetOutput(output, insertBeforeLastWhitespace(output.String(), ":"))
				} else {
					return false, newColonExpectedError(*i)
				}
			}
			processedValue, err := parseValue(text, i, output)
			if err != nil {
				return false, err
			}
			if !processedValue {
				if processedColon || truncatedText {
					// repair missing object value
					output.WriteString("null")
				} else {
					return false, nil
				}
			}
		}

		if *i < len(*text) && (*text)[*i] == codeClosingBrace {
			output.WriteRune((*text)[*i])
			*i++
		} else {
			// repair missing end bracket
			resetOutput(output, insertBeforeLastWhitespace(output.String(), "}"))
		}
		return true, nil
	}
	return false, nil
}

// parseArray parses an array from the input text.
// Returns (success, error) where error is non-nil for non-repairable issues.
func parseArray(text *[]rune, i *int, output *strings.Builder) (bool, error) {
	if *i >= len(*text) {
		return false, nil
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
					// repair missing comma
					resetOutput(output, insertBeforeLastWhitespace(output.String()[:oBefore], ","))
				}
			} else {
				initial = false
			}

			skipEllipsis(text, i, output)

			processedValue, err := parseValue(text, i, output)
			if err != nil {
				return false, err
			}

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
				if outputStrWithoutCommaQuote, ok := strings.CutSuffix(outputStr, ",\""); ok {
					// Ensure the string contains more than just that comma.
					// The minimal string we do NOT want to alter is ",",
					// which would look like ["\",\"]. That has length 3
					// including the comma and quotes -> 4 characters in the
					// output (opening [, closing ], quotes). A safer check is
					// to verify that inside the quotes we have more than one
					// character.

					// Find the position of the opening quote for this value.
					lastQuote := strings.LastIndex(outputStrWithoutCommaQuote, "\"")
					if lastQuote != -1 && len(outputStrWithoutCommaQuote)-lastQuote > 2 {
						resetOutput(output, outputStrWithoutCommaQuote+"\"")
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
				resetOutput(output, stripLastOccurrence(output.String(), ",", false))
				break
			}
		}

		if *i < len(*text) && (*text)[*i] == codeClosingBracket {
			output.WriteRune((*text)[*i])
			*i++
		} else {
			// repair missing closing array bracket
			resetOutput(output, insertBeforeLastWhitespace(output.String(), "]"))
		}
		return true, nil
	}
	return false, nil
}

// parseNewlineDelimitedJSON parses Newline Delimited JSON (NDJSON) from the input text.
func parseNewlineDelimitedJSON(text *[]rune, i *int, output *strings.Builder) {
	for {
		processedValue, err := parseValue(text, i, output)
		if err != nil || !processedValue {
			resetOutput(output, stripLastOccurrence(output.String(), ",", false))
			break
		}

		if !parseCharacter(text, i, output, codeComma) {
			resetOutput(output, insertBeforeLastWhitespace(output.String(), ","))
		}
	}

	resetOutput(output, "[\n"+output.String()+"\n]")
}
