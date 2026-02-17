package jsonrepair

import (
	"path/filepath"
	"regexp"
	"strings"
)

// prevNonWhitespaceIndex returns the index of the last non-whitespace rune
// at or before start. Returns -1 if not found.
func prevNonWhitespaceIndex(text []rune, start int) int {
	for i := start; i >= 0; i-- {
		if !isWhitespace(text[i]) {
			return i
		}
	}
	return -1
}

// atEndOfNumber checks if the end of a number has been reached in the input text.
func atEndOfNumber(text *[]rune, i *int) bool {
	return *i >= len(*text) || isDelimiter((*text)[*i]) || isWhitespace((*text)[*i])
}

// repairNumberEndingWithNumericSymbol repairs numbers cut off at the end.
func repairNumberEndingWithNumericSymbol(text *[]rune, start int, i *int, output *strings.Builder) {
	output.WriteString(string((*text)[start:*i]))
	output.WriteByte('0')
}

// stripLastOccurrence removes the last occurrence of substr from text.
// If stripRemaining is true, removes everything from the match onwards.
func stripLastOccurrence(text, substr string, stripRemaining bool) string {
	index := strings.LastIndex(text, substr)
	if index == -1 {
		return text
	}
	if stripRemaining {
		return text[:index]
	}
	return text[:index] + text[index+len(substr):]
}

// insertBeforeLastWhitespace inserts text before trailing whitespace.
// If no trailing whitespace exists, appends to the end.
func insertBeforeLastWhitespace(s, text string) string {
	if len(s) == 0 || !isWhitespace(rune(s[len(s)-1])) {
		return s + text
	}

	index := len(s) - 1
	for index >= 0 && isWhitespace(rune(s[index])) {
		index--
	}

	return s[:index+1] + text + s[index+1:]
}

// removeAtIndex removes a substring from the input text at a specific index.
func removeAtIndex(text string, start, count int) string {
	return text[:start] + text[start+count:]
}

// isHex checks if a rune is a hexadecimal digit.
func isHex(c rune) bool {
	return (c >= codeZero && c <= codeNine) ||
		(c >= codeUppercaseA && c <= codeUppercaseF) ||
		(c >= codeLowercaseA && c <= codeLowercaseF)
}

// isDigit checks if a rune is a digit.
func isDigit(c rune) bool {
	return c >= codeZero && c <= codeNine
}

// isValidStringCharacter checks if a character is valid inside a JSON string.
// Valid characters are >= U+0020 (space).
func isValidStringCharacter(c rune) bool {
	return c >= 0x0020
}

// isDelimiter checks if a character is a delimiter.
func isDelimiter(c rune) bool {
	return c == ',' || c == ':' || c == '[' || c == ']' || c == '/' ||
		c == '{' || c == '}' || c == '(' || c == ')' || c == '\n' || c == '+'
}

// isStartOfValue checks if a rune is the start of a JSON value.
func isStartOfValue(c rune) bool {
	if c == '{' || c == '[' || c == '_' || c == '-' || isQuote(c) {
		return true
	}
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// isControlCharacter checks if a rune is a control character.
func isControlCharacter(c rune) bool {
	return c == codeNewline ||
		c == codeReturn ||
		c == codeTab ||
		c == codeBackspace ||
		c == codeFormFeed
}

// isWhitespace checks if a rune is a whitespace character.
func isWhitespace(c rune) bool {
	return c == codeSpace ||
		c == codeNewline ||
		c == codeTab ||
		c == codeReturn
}

// isSpecialWhitespace checks if a rune is a special whitespace character.
func isSpecialWhitespace(c rune) bool {
	return c == codeNonBreakingSpace ||
		(c >= codeEnQuad && c <= codeHairSpace) ||
		c == codeNarrowNoBreakSpace ||
		c == codeMediumMathematicalSpace ||
		c == codeIdeographicSpace
}

// isQuote checks if a rune is a quote character.
func isQuote(c rune) bool {
	return isDoubleQuoteLike(c) || isSingleQuoteLike(c)
}

// isDoubleQuoteLike checks if a rune is a double quote or variant.
func isDoubleQuoteLike(c rune) bool {
	return c == codeDoubleQuote ||
		c == codeDoubleQuoteLeft ||
		c == codeDoubleQuoteRight
}

// isDoubleQuote checks if a rune is a double quote.
func isDoubleQuote(c rune) bool {
	return c == codeDoubleQuote
}

// isSingleQuoteLike checks if a rune is a single quote or variant.
func isSingleQuoteLike(c rune) bool {
	return c == codeQuote ||
		c == codeQuoteLeft ||
		c == codeQuoteRight ||
		c == codeGraveAccent ||
		c == codeAcuteAccent
}

// isSingleQuote checks if a rune is a single quote.
func isSingleQuote(c rune) bool {
	return c == codeQuote
}

// endsWithCommaOrNewline checks if the string ends with a comma or newline.
// Only matches commas outside of quoted strings.
func endsWithCommaOrNewline(text string) bool {
	if len(text) == 0 {
		return false
	}

	runes := []rune(text)
	i := len(runes) - 1

	// Skip trailing whitespace (excluding newlines)
	for i >= 0 && (runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\r') {
		i--
	}

	if i < 0 {
		return false
	}

	if runes[i] != ',' && runes[i] != '\n' {
		return false
	}

	// If text ends with a quote, use regex to verify comma is outside string
	trimmed := strings.TrimSpace(text)
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '"' {
		return endsWithCommaOrNewlineRe.MatchString(text)
	}

	return true
}

// isFunctionNameCharStart checks if a rune is a valid function name start character.
func isFunctionNameCharStart(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '$'
}

// isFunctionNameChar checks if a rune is a valid function name character.
func isFunctionNameChar(c rune) bool {
	return isFunctionNameCharStart(c) || isDigit(c)
}

// isUnquotedStringDelimiter checks if a character is a delimiter for unquoted strings.
// Similar to isDelimiter but excludes ':' since colons are allowed inside
// unquoted values until a key/value separator is detected.
func isUnquotedStringDelimiter(c rune) bool {
	return c == ',' || c == '[' || c == ']' || c == '/' ||
		c == '{' || c == '}' || c == '\n' || c == '+'
}

// isWhitespaceExceptNewline checks if a rune is whitespace excluding newline.
func isWhitespaceExceptNewline(c rune) bool {
	return c == codeSpace || c == codeTab || c == codeReturn
}

// URL-related regular expressions.
var (
	regexURLStart = regexp.MustCompile(`^(https?|ftp|mailto|file|data|irc)://`)
)

// isURLChar checks if a rune is a valid URL character.
func isURLChar(c rune) bool {
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
		return true
	}
	return c == '-' || c == '.' || c == '_' || c == '~' || c == ':' || c == '/' ||
		c == '?' || c == '#' || c == '@' || c == '!' || c == '$' || c == '&' ||
		c == '\'' || c == '(' || c == ')' || c == '*' || c == '+' || c == ';' || c == '='
}

// Regular expression cache for improved performance.
var (
	leadingZeroRe            = regexp.MustCompile(`^0\d`)
	endsWithCommaOrNewlineRe = regexp.MustCompile(`"[ \t\r]*[,\n][ \t\r]*$`)
	driveLetterRe            = regexp.MustCompile(`^[A-Za-z]:\\`)
	containsDriveRe          = regexp.MustCompile(`[A-Za-z]:\\`)
	base64Re                 = regexp.MustCompile(`^[A-Za-z0-9+/=]{20,}$`)
	fileExtensionRe          = regexp.MustCompile(`(?i)\.[a-z0-9]{2,5}(\?|$|\\|"|/)`)
	unicodeEscapeRe          = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	urlEncodingRe            = regexp.MustCompile(`%[0-9a-fA-F]{2}`)
)

// ================================
// PATH PATTERN CONSTANTS
// ================================

// windowsPathPatterns contains common Windows directory patterns for path detection.
var windowsPathPatterns = []string{
	// System directories
	"program files", "system32", "windows\\", "programdata",
	// User directories
	"users\\", "documents", "desktop", "downloads", "music", "pictures", "videos", "appdata", "roaming", "public",
	// System functional directories
	"temp\\", "fonts", "startup", "sendto", "recent", "nethood", "cookies", "cache", "history", "favorites", "templates",
	// Web and development directories
	"inetpub", "wwwroot", "node_modules", "npm",
}

// unixPathPatterns contains common Unix/macOS directory patterns for path detection.
var unixPathPatterns = []string{
	// Standard Unix directories
	"/bin/", "/etc/", "/var/", "/usr/", "/opt/", "/home/", "/tmp/", "/lib/", "/lib64/",
	// System directories
	"/proc/", "/dev/", "/sys/", "/run/", "/srv/", "/mnt/", "/media/", "/boot/", "/snap/",
	// Application and data directories
	"/usr/share/", "/usr/local/", "/usr/src/", "/var/log/", "/var/lib/", "/var/cache/", "/var/spool/",
	// macOS specific directories
	"/Applications/", "/Library/", "/System/", "/Users/",
}

// commonFileExtensions contains file extensions commonly found in file paths.
var commonFileExtensions = []string{
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
	".mp3", ".mp4", ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico",
	// Data files
	".dat", ".bin", ".raw", ".dump",
}

// ================================
// EARLY EXCLUSION FILTERS
// ================================

// hasExcessiveEscapeSequences checks if content has too many escape sequences to be a valid file path.
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
	for i := range len(content) - 1 {
		if content[i] == '\\' {
			next := content[i+1]
			if next == 'n' || next == 't' || next == 'r' || next == 'b' || next == 'f' || next == '"' || next == '\\' {
				escapeCount++
			}
		}
	}

	// If more than 30% of content is escape sequences, likely not a path
	return escapeCount > 0 && float64(escapeCount*2)/float64(len(content)) > 0.3
}

// isLikelyTextBlob identifies content that has text-like characteristics.
func isLikelyTextBlob(content string) bool {
	if len(content) < 3 {
		return false
	}

	// Multiple consecutive spaces (rare in paths)
	if strings.Contains(content, "  ") {
		return true
	}

	// Contains line breaks or tabs
	if strings.ContainsAny(content, "\n\t\r") {
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

// isBase64String checks if content appears to be base64 encoded.
func isBase64String(content string) bool {
	if len(content) < 20 {
		return false
	}
	return base64Re.MatchString(content)
}

// hasURLEncoding checks if content contains URL encoding patterns.
func hasURLEncoding(content string) bool {
	return urlEncodingRe.MatchString(content)
}

// ================================
// PATH FORMAT DETECTION
// ================================

// isWindowsAbsolutePath checks for Windows absolute paths (drive letter format).
func isWindowsAbsolutePath(content string) bool {
	return driveLetterRe.MatchString(content) || containsDriveRe.MatchString(content)
}

// isUNCPath checks for UNC (Universal Naming Convention) paths.
func isUNCPath(content string) bool {
	if !strings.HasPrefix(content, `\\`) || strings.HasPrefix(content, `\\\\`) {
		return false
	}
	parts := strings.Split(content, `\`)
	return len(parts) >= 4 && len(parts[2]) > 0 && len(parts[3]) > 0
}

// isUnixAbsolutePath checks for Unix absolute paths.
func isUnixAbsolutePath(content string) bool {
	return strings.HasPrefix(content, "/") || strings.HasPrefix(content, "~/")
}

// isURLPath checks for URL-style file paths.
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
			return hasValidPathStructure(pathPart[slashIndex:])
		}
	}

	return false
}

// ================================
// STRUCTURAL VALIDATION
// ================================

// containsPathSeparator checks if content contains valid path separators.
func containsPathSeparator(content string) bool {
	return strings.Contains(content, "/") || strings.Contains(content, "\\")
}

// countValidPathSegments counts meaningful path segments.
func countValidPathSegments(content, separator string) int {
	parts := strings.Split(content, separator)
	count := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 && part != "." && part != ".." {
			count++
		}
	}

	return count
}

// hasFileExtension checks if content has a valid file extension.
func hasFileExtension(content string) bool {
	// Use Go's filepath.Ext for standard detection
	ext := filepath.Ext(content)
	if len(ext) > 1 && len(ext) <= 6 {
		return true
	}

	// Use regex for additional patterns
	return fileExtensionRe.MatchString(content)
}

// hasValidPathStructure validates the overall path structure.
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

	// Windows common directories - reuse package-level patterns
	for _, pattern := range windowsPathPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	// Unix system directories
	if strings.HasPrefix(pathStr, "/") {
		for _, dir := range unixPathPatterns {
			if strings.Contains(lowerPath, dir) {
				return true
			}
		}
	}

	return false
}

// isValidPathCharacter checks if a character is valid in file paths.
func isValidPathCharacter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
		return true
	}
	return r == '/' || r == '\\' || r == ':' || r == '.' ||
		r == '-' || r == '_' || r == ' ' || r == '~'
}

// hasReasonableCharacterDistribution checks character distribution for path-like content.
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

// matchesWindowsPathPattern checks if content matches common Windows directory patterns.
func matchesWindowsPathPattern(lowerContent, content string) bool {
	for _, pattern := range windowsPathPatterns {
		if strings.Contains(lowerContent, pattern) && containsPathSeparator(content) {
			return true
		}
	}
	return false
}

// matchesUnixPathPattern checks if content matches common Unix/macOS directory patterns.
func matchesUnixPathPattern(lowerContent string) bool {
	for _, pattern := range unixPathPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

// hasCommonFileExtension checks if content ends with a common file extension.
func hasCommonFileExtension(lowerContent string) bool {
	for _, ext := range commonFileExtensions {
		if strings.HasSuffix(lowerContent, ext) {
			return true
		}
	}
	return false
}

// isExcludedURL checks if content is a URL that should be excluded from path detection.
func isExcludedURL(lowerContent, content string) bool {
	if strings.HasPrefix(lowerContent, "http://") || strings.HasPrefix(lowerContent, "https://") {
		return true
	}
	return strings.HasPrefix(lowerContent, "ftp://") && len(content) > 6 && !strings.Contains(content[6:], "/")
}

// passesEarlyExclusionFilters checks if content passes all early exclusion filters.
func passesEarlyExclusionFilters(content string) bool {
	return !hasExcessiveEscapeSequences(content) &&
		!isLikelyTextBlob(content) &&
		!isBase64String(content) &&
		!hasURLEncoding(content)
}

// matchesAbsolutePathFormat checks if content matches any absolute path format.
func matchesAbsolutePathFormat(content string) bool {
	return isURLPath(content) ||
		isWindowsAbsolutePath(content) ||
		isUNCPath(content) ||
		isUnixAbsolutePath(content)
}

// isLikelyFilePath determines if a string looks like a file path.
func isLikelyFilePath(content string) bool {
	if len(content) < 2 {
		return false
	}

	lowerContent := strings.ToLower(content)

	// Early URL exclusions
	if isExcludedURL(lowerContent, content) {
		return false
	}

	// Early exclusion filters
	if !passesEarlyExclusionFilters(content) {
		return false
	}

	// Format-specific detection (high confidence)
	if matchesAbsolutePathFormat(content) {
		return true
	}

	// Check for common Windows directory patterns
	if matchesWindowsPathPattern(lowerContent, content) {
		return true
	}

	// Check for Unix system directory patterns
	if strings.Contains(content, "/") && matchesUnixPathPattern(lowerContent) {
		return true
	}

	// Structural validation for relative paths
	if !containsPathSeparator(content) {
		return false
	}

	// Check for common file extensions
	if hasFileExtension(content) && hasCommonFileExtension(lowerContent) {
		return true
	}

	if !hasReasonableCharacterDistribution(content) {
		return false
	}

	return hasValidPathStructure(content)
}

// analyzePotentialFilePath analyzes text to determine if it contains file paths.
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
					for j := range 6 {
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
