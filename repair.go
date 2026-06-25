// Package jsonrepair repairs malformed JSON strings returned by LLMs,
// JavaScript snippets, and other JSON-like sources.
//
// Public APIs return errors instead of panicking, and the package does not expose
// Must* helpers. The only panic-capable paths are package initialization of
// internal regexp.MustCompile literals, which can fail only if the package source
// contains an invalid regular expression.
package jsonrepair

import (
	"fmt"
	"strings"
)

// Repair repairs malformed JSON-like input into valid JSON.
//
// It returns a structured [*Error] when the input is empty or cannot be repaired.
func Repair(text string) (string, error) {
	if len(text) == 0 {
		return "", newUnexpectedEndError(0)
	}

	runes := []rune(text)
	i := 0
	var output strings.Builder

	parseMarkdownCodeBlock(&runes, &i, []string{"```", "[```", "{```"}, &output)

	success, err := parseValue(&runes, &i, &output)
	if err != nil {
		return "", err
	}
	if !success {
		return "", newUnexpectedEndError(len(runes))
	}

	parseMarkdownCodeBlock(&runes, &i, []string{"```", "```]", "```}"}, &output)

	processedComma := parseCharacter(&runes, &i, &output, codeComma)
	if processedComma {
		parseWhitespaceAndSkipComments(&runes, &i, &output, true)
	}

	if i < len(runes) && isStartOfValue(runes[i]) && endsWithCommaOrNewline(output.String()) {
		if !processedComma {
			resetOutput(&output, insertBeforeLastWhitespace(output.String(), ","))
		}
		parseNewlineDelimitedJSON(&runes, &i, &output)
	} else if processedComma {
		resetOutput(&output, stripLastOccurrence(output.String(), ",", false))
	}

	for i < len(runes) && (runes[i] == codeClosingBrace || runes[i] == codeClosingBracket) {
		i++
		parseWhitespaceAndSkipComments(&runes, &i, &output, true)
	}

	parseWhitespaceAndSkipComments(&runes, &i, &output, true)

	if i >= len(runes) {
		return output.String(), nil
	}

	message := fmt.Sprintf("unexpected character %q", string(runes[i]))
	return "", newUnexpectedCharacterError(message, i)
}

// parseValue determines the type of the next value and parses it.
// Returns (success, error) where error is non-nil for non-repairable issues.
func parseValue(text *[]rune, i *int, output *strings.Builder) (bool, error) {
	parseWhitespaceAndSkipComments(text, i, output, true)

	processedObj, err := parseObject(text, i, output)
	if err != nil {
		return false, err
	}
	if processedObj {
		parseWhitespaceAndSkipComments(text, i, output, true)
		return true, nil
	}

	processed, err := parseArray(text, i, output)
	if err != nil {
		return false, err
	}
	if processed {
		parseWhitespaceAndSkipComments(text, i, output, true)
		return true, nil
	}

	stringProcessed, err := parseString(text, i, output, false, -1)
	if err != nil {
		return false, err
	}
	processed = stringProcessed ||
		parseNumber(text, i, output) ||
		parseKeywords(text, i, output)
	if !processed {
		processed, err = parseUnquotedStringWithMode(text, i, output, false)
		if err != nil {
			return false, err
		}
	}
	if !processed {
		processed = parseRegex(text, i, output)
	}

	parseWhitespaceAndSkipComments(text, i, output, true)
	return processed, nil
}
