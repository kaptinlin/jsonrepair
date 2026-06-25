package jsonrepair

import "strings"

// parseMarkdownCodeBlock parses and skips Markdown fenced code blocks like ``` or ```json.
func parseMarkdownCodeBlock(text *[]rune, i *int, blocks []string, output *strings.Builder) bool {
	if !skipMarkdownCodeBlock(text, i, blocks, output) {
		return false
	}

	if *i < len(*text) && isFunctionNameCharStart((*text)[*i]) {
		// Strip the optional language specifier like "json"
		for *i < len(*text) && isFunctionNameChar((*text)[*i]) {
			*i++
		}
	}

	parseWhitespace(text, i, output, true)

	return true
}

// skipMarkdownCodeBlock checks if we're at a Markdown code block marker and skips it.
func skipMarkdownCodeBlock(text *[]rune, i *int, blocks []string, output *strings.Builder) bool {
	parseWhitespace(text, i, output, true)

	for _, block := range blocks {
		end := *i + len(block)
		if end > len(*text) {
			continue
		}
		if string((*text)[*i:end]) == block {
			*i = end
			return true
		}
	}
	return false
}
