package jsonrepair

import (
	"path/filepath"
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

// isValidStringCharacter checks if a character is valid inside a JSON string
// Matches TypeScript version: char >= '\u0020'
func isValidStringCharacter(char rune) bool {
	return char >= 0x0020
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
var regexURLStart = regexp.MustCompile(`^(https?|ftp|mailto|file|data|irc)://`)
var regexURLChar = regexp.MustCompile(`^[A-Za-z0-9\-._~:/?#@!$&'()*+;=]$`)

// isURLChar checks if a rune is a valid URL character.
func isURLChar(code rune) bool {
	return regexURLChar.MatchString(string(code))
}

// Regular expression cache for improved performance
var (
	driveLetterRe   = regexp.MustCompile(`^[A-Za-z]:\\`)
	containsDriveRe = regexp.MustCompile(`[A-Za-z]:\\`)
	base64Re        = regexp.MustCompile(`^[A-Za-z0-9+/=]{20,}$`)
	fileExtensionRe = regexp.MustCompile(`(?i)\.[a-z0-9]{2,5}(\?|$|\\|"|/)`)
	unicodeEscapeRe = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	urlEncodingRe   = regexp.MustCompile(`%[0-9a-fA-F]{2}`)
)

// ================================
// EARLY EXCLUSION FILTERS
// ================================

// hasExcessiveEscapeSequences checks if content has too many escape sequences to be a valid file path
func hasExcessiveEscapeSequences(content string) bool {
	if len(content) < 3 {
		return false
	}

	// Count Unicode escape sequences
	unicodeMatches := unicodeEscapeRe.FindAllString(content, -1)
	if len(unicodeMatches) >= 2 {
		totalUnicodeLength := len(unicodeMatches) * 6 // Each \uXXXX is 6 chars
		if float64(totalUnicodeLength)/float64(len(content)) > 0.6 {
			return true
		}
	}

	// Count general escape sequences
	escapeCount := 0
	for i := 0; i < len(content)-1; i++ {
		if content[i] == '\\' {
			next := content[i+1]
			if next == 'n' || next == 't' || next == 'r' || next == 'b' || next == 'f' || next == '"' || next == '\\' {
				escapeCount++
			}
		}
	}

	// If more than 30% of content is escape sequences, likely not a path
	if escapeCount > 0 && float64(escapeCount*2)/float64(len(content)) > 0.3 {
		return true
	}

	return false
}

// isLikelyTextBlob identifies content that has text-like characteristics
func isLikelyTextBlob(content string) bool {
	if len(content) < 3 {
		return false
	}

	// Multiple consecutive spaces (rare in paths)
	if strings.Contains(content, "  ") {
		return true
	}

	// Contains line breaks or tabs
	if strings.Contains(content, "\n") || strings.Contains(content, "\t") || strings.Contains(content, "\r") {
		return true
	}

	// Sentence-like punctuation patterns
	if strings.Contains(content, ". ") || strings.Contains(content, "! ") || strings.Contains(content, "? ") {
		return true
	}

	// Too many spaces for a typical path (more than 5 spaces instead of 3)
	spaceCount := strings.Count(content, " ")
	if spaceCount > 5 {
		return true
	}

	// Sentence-like capitalization pattern (more restrictive)
	if len(content) > 10 && content[0] >= 'A' && content[0] <= 'Z' && spaceCount > 2 {
		lowercaseAfterSpace := 0
		foundSpace := false
		for _, r := range content[1:] {
			if r == ' ' {
				foundSpace = true
			} else if foundSpace && r >= 'a' && r <= 'z' {
				lowercaseAfterSpace++
			}
		}
		if lowercaseAfterSpace >= 3 {
			return true
		}
	}

	return false
}

// isBase64String checks if content appears to be base64 encoded
func isBase64String(content string) bool {
	if len(content) < 20 {
		return false
	}
	return base64Re.MatchString(content)
}

// hasURLEncoding checks if content contains URL encoding patterns
func hasURLEncoding(content string) bool {
	return urlEncodingRe.MatchString(content)
}

// ================================
// PATH FORMAT DETECTION
// ================================

// isWindowsAbsolutePath checks for Windows absolute paths (drive letter format)
func isWindowsAbsolutePath(content string) bool {
	return driveLetterRe.MatchString(content) || containsDriveRe.MatchString(content)
}

// isUNCPath checks for UNC (Universal Naming Convention) paths
func isUNCPath(content string) bool {
	if !strings.HasPrefix(content, `\\`) || strings.HasPrefix(content, `\\\\`) {
		return false
	}

	parts := strings.Split(content, `\`)
	// UNC: \\server\share\path... (parts[0]="", parts[1]="", parts[2]=server, parts[3]=share)
	return len(parts) >= 4 && len(parts[2]) > 0 && len(parts[3]) > 0
}

// isUnixAbsolutePath checks for Unix absolute paths
func isUnixAbsolutePath(content string) bool {
	return strings.HasPrefix(content, "/") || strings.HasPrefix(content, "~/")
}

// isURLPath checks for URL-style file paths
func isURLPath(content string) bool {
	lowerContent := strings.ToLower(content)

	// Exclude HTTP/HTTPS URLs
	if strings.HasPrefix(lowerContent, "http://") || strings.HasPrefix(lowerContent, "https://") {
		return false
	}

	// File protocol
	if strings.HasPrefix(lowerContent, "file://") {
		pathPart := content[7:]
		return len(pathPart) > 1 && hasValidPathStructure(pathPart)
	}

	// SMB/CIFS protocol
	if strings.HasPrefix(lowerContent, "smb://") {
		pathPart := content[6:]
		return len(pathPart) > 1 && hasValidPathStructure(pathPart)
	}

	// FTP with file path
	if strings.HasPrefix(lowerContent, "ftp://") {
		pathPart := content[6:]
		if slashIndex := strings.Index(pathPart, "/"); slashIndex > 0 {
			actualPath := pathPart[slashIndex:]
			return hasValidPathStructure(actualPath)
		}
	}

	return false
}

// ================================
// STRUCTURAL VALIDATION
// ================================

// containsPathSeparator checks if content contains valid path separators
func containsPathSeparator(content string) bool {
	return strings.Contains(content, "/") || strings.Contains(content, "\\")
}

// countValidPathSegments counts meaningful path segments
func countValidPathSegments(content string, separator string) int {
	parts := strings.Split(content, separator)
	meaningfulParts := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 && part != "." && part != ".." {
			meaningfulParts++
		}
	}

	return meaningfulParts
}

// hasFileExtension checks if content has a valid file extension
func hasFileExtension(content string) bool {
	// Use Go's filepath.Ext for standard detection
	ext := filepath.Ext(content)
	if len(ext) > 1 && len(ext) <= 6 {
		return true
	}

	// Use regex for additional patterns
	return fileExtensionRe.MatchString(content)
}

// hasValidPathStructure validates the overall path structure
func hasValidPathStructure(pathStr string) bool {
	if len(pathStr) < 2 {
		return false
	}

	// Check for path separators
	if !containsPathSeparator(pathStr) {
		return false
	}

	// Determine separator type
	separator := "/"
	if strings.Contains(pathStr, "\\") {
		separator = "\\"
	}

	// Count meaningful segments
	meaningfulParts := countValidPathSegments(pathStr, separator)
	if meaningfulParts < 2 {
		return false
	}

	// Check for file extension (optional but helpful)
	hasExt := hasFileExtension(pathStr)

	// More lenient requirements:
	// - If has extension, accept with 2+ parts
	// - If no extension, require 3+ parts OR known path patterns
	if hasExt {
		return true
	}

	// For paths without extensions, be more lenient
	if meaningfulParts >= 3 {
		return true
	}

	// Special cases for known path patterns
	lowerPath := strings.ToLower(pathStr)

	// Windows common directories
	windowsDirs := []string{
		"program files", "windows", "users", "temp", "system32", "documents", "programdata",
		"desktop", "downloads", "music", "pictures", "videos", "appdata", "roaming", "public",
		"inetpub", "wwwroot", "node_modules", "npm",
	}
	for _, dir := range windowsDirs {
		if strings.Contains(lowerPath, dir) {
			return true
		}
	}

	// Unix system directories
	if strings.HasPrefix(pathStr, "/") {
		unixDirs := []string{
			"/bin/", "/etc/", "/var/", "/usr/", "/opt/", "/home/", "/tmp/", "/lib/",
			"/proc/", "/dev/", "/sys/", "/run/", "/srv/", "/mnt/", "/media/", "/boot/",
			"/Applications/", "/Library/", "/System/", "/Users/",
		}
		for _, dir := range unixDirs {
			if strings.Contains(lowerPath, dir) {
				return true
			}
		}
	}

	return false
}

// isValidPathCharacter checks if a character is valid in file paths
func isValidPathCharacter(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '/' || r == '\\' || r == ':' || r == '.' ||
		r == '-' || r == '_' || r == ' ' || r == '~'
}

// hasReasonableCharacterDistribution checks character distribution for path-like content
func hasReasonableCharacterDistribution(content string) bool {
	if len(content) == 0 {
		return false
	}

	validChars := 0
	for _, r := range content {
		if isValidPathCharacter(r) {
			validChars++
		}
	}

	// At least 70% of characters should be valid path characters
	return float64(validChars)/float64(len(content)) >= 0.7
}

// ================================
// MAIN PATH DETECTION
// ================================

// isLikelyFilePath determines if a string content looks like a file path
// using a structured, layer-based approach
func isLikelyFilePath(content string) bool {
	if len(content) < 2 {
		return false
	}

	// EARLY STRONG EXCLUSIONS: HTTP/HTTPS URLs
	lowerContent := strings.ToLower(content)
	if strings.HasPrefix(lowerContent, "http://") || strings.HasPrefix(lowerContent, "https://") {
		return false
	}

	// Early exclude FTP URLs without file paths
	if strings.HasPrefix(lowerContent, "ftp://") && !strings.Contains(content[6:], "/") {
		return false
	}

	// Early exclusion filters
	if hasExcessiveEscapeSequences(content) {
		return false
	}

	if isLikelyTextBlob(content) {
		return false
	}

	if isBase64String(content) {
		return false
	}

	if hasURLEncoding(content) {
		return false
	}

	// Format-specific detection (high confidence)
	if isURLPath(content) {
		return true
	}

	if isWindowsAbsolutePath(content) {
		return true
	}

	if isUNCPath(content) {
		return true
	}

	if isUnixAbsolutePath(content) {
		return true
	}

	// Additional pattern detection for common paths
	// Check for common Windows directory patterns
	windowsPatterns := []string{
		// System directories
		"program files", "system32", "windows\\", "programdata",
		// User directories
		"users\\", "documents", "desktop", "downloads", "music", "pictures", "videos", "appdata", "roaming", "public",
		// System functional directories
		"temp\\", "fonts", "startup", "sendto", "recent", "nethood", "cookies", "cache", "history", "favorites", "templates",
	}
	for _, pattern := range windowsPatterns {
		if strings.Contains(lowerContent, pattern) && containsPathSeparator(content) {
			return true
		}
	}

	// Check for Unix system directory patterns
	if strings.Contains(content, "/") {
		unixPatterns := []string{
			// Standard Unix directories
			"/bin/", "/etc/", "/var/", "/usr/", "/opt/", "/home/", "/tmp/", "/lib/", "/lib64/",
			// System directories
			"/proc/", "/dev/", "/sys/", "/run/", "/srv/", "/mnt/", "/media/", "/boot/", "/snap/",
			// Application and data directories
			"/usr/share/", "/usr/local/", "/usr/src/", "/var/log/", "/var/lib/", "/var/cache/", "/var/spool/",
			// macOS specific directories
			"/Applications/", "/Library/", "/System/", "/Users/",
		}
		for _, pattern := range unixPatterns {
			if strings.Contains(lowerContent, pattern) {
				return true
			}
		}
	}

	// Structural validation for relative paths
	if !containsPathSeparator(content) {
		return false
	}

	// Relaxed check for simple backup/config files with common extensions
	if hasFileExtension(content) {
		commonFileExts := []string{
			// Configuration files
			".config", ".cfg", ".ini", ".conf", ".properties", ".toml",
			// Data formats
			".json", ".xml", ".yml", ".yaml", ".csv", ".tsv",
			// Backup and temporary files
			".backup", ".bak", ".old", ".tmp", ".temp", ".swp", ".~",
			// Log and debug files
			".log", ".out", ".err", ".debug", ".trace",
			// Database files
			".db", ".sqlite", ".sqlite3", ".mdb",
			// Document files
			".txt", ".md", ".readme", ".doc", ".docx", ".pdf",
			// Archive files
			".zip", ".tar", ".gz", ".rar", ".7z", ".bz2", ".xz",
			// Code files
			".js", ".ts", ".py", ".go", ".java", ".cpp", ".c", ".h", ".cs", ".php", ".rb", ".rs",
			// Media files
			".mp3", ".mp4", ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico", ".mp3", ".mp4", ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico",
			// Data files
			".dat", ".bin", ".raw", ".dump",
		}
		for _, ext := range commonFileExts {
			if strings.HasSuffix(lowerContent, ext) {
				return true
			}
		}
	}

	if !hasReasonableCharacterDistribution(content) {
		return false
	}

	return hasValidPathStructure(content)
}

// analyzePotentialFilePath analyzes a portion of text to determine if it contains file paths
// This function has been optimized for structural detection
func analyzePotentialFilePath(text *[]rune, startPos int) bool {
	if startPos >= len(*text) || (*text)[startPos] != '"' {
		return false
	}

	// Extract string content
	i := startPos + 1
	var contentBuilder strings.Builder
	hasPathSeparator := false

	// Collect content until closing quote (with reasonable limit)
	for i < len(*text) && i < startPos+150 {
		char := (*text)[i]

		if char == '"' {
			break
		}

		// Track path separators
		if char == '\\' || char == '/' {
			hasPathSeparator = true
		}

		// Handle escape sequences for path detection
		if char == '\\' && i+1 < len(*text) {
			nextChar := (*text)[i+1]
			switch nextChar {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				// Preserve escape sequences as-is for path analysis
				contentBuilder.WriteRune(char)
				contentBuilder.WriteRune(nextChar)
				i += 2
				continue
			case 'u':
				// Unicode escape
				if i+5 < len(*text) {
					for j := 0; j < 6; j++ {
						contentBuilder.WriteRune((*text)[i+j])
					}
					i += 6
					continue
				}
			}
		}

		contentBuilder.WriteRune(char)
		i++
	}

	content := contentBuilder.String()

	// Pre-validation checks
	if len(content) < 3 {
		return false
	}

	if !hasPathSeparator {
		return false
	}

	return isLikelyFilePath(content)
}
