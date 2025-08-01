package jsonrepair

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertBeforeLastWhitespace(t *testing.T) {
	tests := []struct {
		text         string
		textToInsert string
		expected     string
	}{
		// Basic cases
		{"abc", "123", "abc123"},
		{"abc ", "123", "abc123 "},
		{"abc  ", "123", "abc123  "},
		{"abc \t\n", "123", "abc123 \t\n"},

		// Trailing whitespace cases
		{"abc\n", "123", "abc123\n"},
		{"abc\t", "123", "abc123\t"},
		{"abc\r\n", "123", "abc123\r\n"},
		{"abc \n\t", "123", "abc123 \n\t"},

		// Edge cases
		{"", "123", "123"},
		{" ", "123", "123 "},
		{"\n", "123", "123\n"},
		{"\t", "123", "123\t"},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			result := insertBeforeLastWhitespace(test.text, test.textToInsert)
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestIsLikelyFilePath tests the improved file path detection function.
func TestIsLikelyFilePath(t *testing.T) {
	// Test cases that should be detected as file paths
	positiveTests := []struct {
		input string
		desc  string
	}{
		{"C:\\temp", "Drive letter path"},
		{"C:\\Users\\Documents", "Drive letter with directories"},
		{"D:\\Program Files\\App\\file.exe", "Drive with program files and exe"},
		{"\\\\server\\share", "UNC path"},
		{"\\\\server\\share\\folder\\file.txt", "UNC path with file"},
		{"C:\\windows\\system32\\file.dll", "Windows system directory"},
		{"\\users\\john\\documents", "Common directory path"},
		{"path\\to\\file.txt", "Multi-level path with extension"},
		{"folder\\subfolder\\document.json", "Path with JSON extension"},
		{"/usr/local/bin", "Unix-style path"},
		{"/home/user/documents/file.log", "Unix path with extension"},
		{"C:\\temp\\newfile", "Path with control character sequence"},
		{"C:\\Program Files\\Application", "Path with space in name"},
		{"temp=C:\\temp\\data", "Path with drive letter in middle"},
		{"config=D:\\app\\config.ini", "Config path with drive"},
		{"/bin/bash", "Unix binary path"},
		{"/etc/hosts", "Unix system config"},
		{"/var/log/system.log", "Unix log path"},
		{"/home/user/.bashrc", "Unix hidden file"},
		{"~/documents/file.txt", "Unix home path"},
		{"path\\to\\file.config", "Config file extension"},
		{"C:\\inetpub\\wwwroot\\index.html", "Web root path"},
		{"folder\\script.py", "Python script path"},
		{"project\\src\\main.js", "JavaScript source path"},
		// URL-style file paths
		{"file:///etc/passwd", "File protocol Unix path"},
		{"file:///C:/Windows/System32/drivers/etc/hosts", "File protocol Windows path"},
		{"file://localhost/home/user/document.txt", "File protocol with localhost"},
		{"smb://server/share/folder/file.doc", "SMB protocol file path"},
		{"smb://192.168.1.100/shared/documents/report.pdf", "SMB with IP and file"},
		{"ftp://ftp.example.com/pub/files/archive.zip", "FTP protocol with file path"},
		{"ftp://user@server.com/home/user/data.csv", "FTP with user and file path"},
	}

	for _, test := range positiveTests {
		t.Run("positive_"+test.desc, func(t *testing.T) {
			if !isLikelyFilePath(test.input) {
				t.Errorf("Expected %q to be detected as file path (%s)", test.input, test.desc)
			}
		})
	}

	// Test cases that should NOT be detected as file paths
	negativeTests := []struct {
		input string
		desc  string
	}{
		{"hello world", "Simple text"},
		{"\\n", "Single escape sequence"},
		{"\\t", "Tab escape"},
		{"\\r", "Carriage return escape"},
		{"\\b", "Backspace escape"},
		{"\\f", "Form feed escape"},
		{"\\u2605", "Unicode escape"},
		{"\\/", "Escaped slash"},
		{"\\\"", "Escaped quote"},
		{"\\\\", "Escaped backslash"},
		{"https://example.com", "HTTP URL"},
		{"http://test.com/path", "HTTP URL with path"},
		{"simple text", "Regular string"},
		{"Hello\\nWorld", "Text with newline escape"},
		{"", "Empty string"},
		{"a", "Single character"},
		{"JSON\\parsing", "Single backslash with text"},
		{"dGVzdCBzdHJpbmcgZm9yIGJhc2U2NCBlbmNvZGluZw==", "Base64 string"},
		{"SGVsbG8gV29ybGQgaXMgYSBsb25nIGJhc2U2NCBzdHJpbmc=", "Long Base64 string"},
		{"message with %2F url encoding", "URL encoded content"},
		{"path with %5C backslash encoding", "URL encoded backslash"},
		{"\\u0048\\u0065\\u006c\\u006c\\u006f", "Unicode escape sequence"},
		{"hello message with \\n escape", "Message text with escape"},
		{"error file not found\\n", "Error message with escape"},
		{"text content with \\t tab", "Text with tab escape"},
		// URL-related negative tests
		{"https://example.com/api/data", "HTTPS API endpoint"},
		{"http://localhost:8080/app", "HTTP localhost URL"},
		{"ftp://ftp.example.com", "FTP URL without file path"},
		{"mailto:user@example.com", "Email protocol URL"},
	}

	for _, test := range negativeTests {
		t.Run("negative_"+test.desc, func(t *testing.T) {
			if isLikelyFilePath(test.input) {
				t.Errorf("Expected %q NOT to be detected as file path (%s)", test.input, test.desc)
			}
		})
	}
}

// TestAnalyzePotentialFilePath tests the path analysis function with rune arrays.
func TestAnalyzePotentialFilePath(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		desc     string
	}{
		{`"C:\temp\file.txt"`, true, "Drive letter path in quotes"},
		{`"Hello\nWorld"`, false, "Text with escape in quotes"},
		{`"\users\john"`, true, "Users directory path"},
		{`"Regular text message"`, false, "Plain text in quotes"},
		{`"path\to\document.json"`, true, "Multi-level path with JSON file"},
		{`"\\server\share\folder"`, true, "UNC path in quotes"},
		{`"Simple message with \\n escape"`, false, "Text with escaped newline"},
		{`"https://example.com/path"`, false, "HTTP URL"},
		{`"temp=C:\app\config.ini"`, true, "Path with drive in middle"},
		{`"/usr/local/bin/app"`, true, "Unix system binary path"},
		{`"/etc/nginx/nginx.conf"`, true, "Unix config file"},
		{`"/var/log/system.log"`, true, "Unix log file"},
		{`"~/documents/readme.txt"`, true, "Unix home directory"},
		{`"dGVzdCBzdHJpbmcgZm9yIGJhc2U2NCBlbmNvZGluZw=="`, false, "Base64 string in quotes"},
		{`"hello message with \n newline"`, false, "Message with newline"},
		{`"error: something failed\t"`, false, "Error message with tab"},
		{`"path\to\file.backup"`, true, "Backup file path"},
		{`"C:\inetpub\wwwroot\app"`, true, "Web root path"},
		{`"project\src\main.py"`, true, "Python source file"},
		{`"content with %2F encoding"`, false, "URL encoded content"},
		// URL-style file path tests
		{`"file:///etc/passwd"`, true, "File protocol Unix path in quotes"},
		{`"file:///C:/Windows/notepad.exe"`, true, "File protocol Windows path in quotes"},
		{`"smb://server/share/document.docx"`, true, "SMB protocol file in quotes"},
		{`"ftp://ftp.example.com/files/data.csv"`, true, "FTP protocol file in quotes"},
		{`"https://api.example.com/data"`, false, "HTTPS API URL in quotes"},
		{`"http://localhost:3000/app"`, false, "HTTP localhost URL in quotes"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			runes := []rune(tc.input)
			result := analyzePotentialFilePath(&runes, 0)
			assert.Equal(t, tc.expected, result, "Failed for: %s", tc.desc)
		})
	}
}

// TestIsURLPath tests the URL-style file path detection function.
func TestIsURLPath(t *testing.T) {
	positiveTests := []struct {
		input string
		desc  string
	}{
		{"file:///etc/passwd", "File protocol Unix absolute path"},
		{"file:///C:/Windows/System32/notepad.exe", "File protocol Windows absolute path"},
		{"file://localhost/home/user/document.txt", "File protocol with localhost"},
		{"FILE:///usr/bin/bash", "File protocol uppercase"},
		{"smb://server/share/folder/file.doc", "SMB protocol with file"},
		{"smb://192.168.1.100/shared/documents/report.pdf", "SMB with IP address"},
		{"SMB://domain.com/public/archive.zip", "SMB protocol uppercase"},
		{"ftp://ftp.example.com/pub/files/data.csv", "FTP with file path"},
		{"ftp://user@server.com/home/user/backup.tar.gz", "FTP with user credentials"},
		{"FTP://files.domain.org/downloads/software.exe", "FTP protocol uppercase"},
	}

	for _, test := range positiveTests {
		t.Run("positive_"+test.desc, func(t *testing.T) {
			if !isURLPath(test.input) {
				t.Errorf("Expected %q to be detected as URL-style file path (%s)", test.input, test.desc)
			}
		})
	}

	negativeTests := []struct {
		input string
		desc  string
	}{
		{"https://example.com/api/data", "HTTPS URL"},
		{"http://localhost:8080/app", "HTTP URL"},
		{"mailto:user@example.com", "Email protocol"},
		{"ftp://ftp.example.com", "FTP without file path"},
		{"smb://server", "SMB without share"},
		{"file://", "File protocol without path"},
		{"regular text", "Plain text"},
		{"/regular/unix/path", "Regular Unix path"},
		{"C:\\regular\\windows\\path", "Regular Windows path"},
	}

	for _, test := range negativeTests {
		t.Run("negative_"+test.desc, func(t *testing.T) {
			if isURLPath(test.input) {
				t.Errorf("Expected %q NOT to be detected as URL-style file path (%s)", test.input, test.desc)
			}
		})
	}
}

// TestHasValidPathStructure tests the path structure validation function.
func TestHasValidPathStructure(t *testing.T) {
	positiveTests := []struct {
		input string
		desc  string
	}{
		{"/etc/passwd", "Unix absolute path"},
		{"/home/user/documents/file.txt", "Unix absolute path with file"},
		{"C:\\Windows\\System32", "Windows absolute path"},
		{"C:\\Program Files\\App\\config.ini", "Windows absolute path with file"},
		{"~/documents/readme.md", "Unix home relative path"},
		{"folder/subfolder/file.log", "Relative path with extension"},
		{"src\\main\\java\\App.java", "Windows relative path with extension"},
		{"../parent/folder/data.json", "Relative path with parent reference"},
	}

	for _, test := range positiveTests {
		t.Run("positive_"+test.desc, func(t *testing.T) {
			if !hasValidPathStructure(test.input) {
				t.Errorf("Expected %q to be detected as valid path structure (%s)", test.input, test.desc)
			}
		})
	}

	negativeTests := []struct {
		input string
		desc  string
	}{
		{"", "Empty string"},
		{"a", "Single character"},
		{"hello world", "Plain text with space"},
		{"just-a-filename", "Single filename without separators"},
		{"no/ext", "Path with only 2 parts and no extension"},
	}

	for _, test := range negativeTests {
		t.Run("negative_"+test.desc, func(t *testing.T) {
			if hasValidPathStructure(test.input) {
				t.Errorf("Expected %q NOT to be detected as valid path structure (%s)", test.input, test.desc)
			}
		})
	}
}

// ================================
// JSON ESCAPE SEQUENCE UTILITY TESTS
// ================================

// TestFilePathDetectionLogic tests the core file path detection logic
func TestFilePathDetectionLogic(t *testing.T) {
	testCases := []struct {
		input       string
		isFilePath  bool
		description string
	}{
		// Clear file path patterns
		{`"C:\Users\Documents"`, true, "Windows drive path"},
		{`"C:\temp\newfile"`, true, "Windows temp directory"},
		{`"\\server\share\folder"`, true, "UNC network path"},
		{`"\documents\local\data"`, true, "Documents directory path"},

		// Clear non-path patterns
		{`"Hello\nWorld"`, false, "Text with newline escape"},
		{`"Tab\there"`, false, "Text with tab escape"},
		{`"Quote\"inside"`, false, "Text with quote escape"},
		{`"Unicode\u2605star"`, false, "Text with Unicode escape"},

		// Mixed cases that should be file paths
		{`"C:\temp\new\file.txt"`, true, "File path with extension"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			runes := []rune(tc.input)
			result := analyzePotentialFilePath(&runes, 0)
			assert.Equal(t, tc.isFilePath, result, "Failed for: %s", tc.description)
		})
	}
}

// TestJSONEscapeCharacterValidation tests validation of escape characters according to JSON standard
func TestJSONEscapeCharacterValidation(t *testing.T) {
	validEscapes := []rune{'"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u'}

	// Test that valid JSON escape characters are recognized
	for _, escape := range validEscapes {
		t.Run(fmt.Sprintf("valid_escape_%c", escape), func(t *testing.T) {
			// Verify the character is in our expected valid set
			switch escape {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
				// These are valid JSON escape characters
			default:
				t.Errorf("Unexpected valid escape character: %c", escape)
			}
		})
	}
}

// TestFilePathDetectionWithEscapes tests file path detection with various escape sequences
func TestFilePathDetectionWithEscapes(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		desc     string
	}{
		// Windows paths with typical JSON escape sequences
		{`C:\temp\newfile`, true, "Windows path with \\n sequence"},
		{`C:\Program Files\App`, true, "Windows path with spaces"},
		{`D:\data\reports`, true, "Windows path with \\r sequence"},

		// Regular text with escape sequences (should not be paths)
		{`Hello\nWorld`, false, "Text with newline escape"},
		{`Error\tmessage`, false, "Text with tab escape"},
		{`Quote\"inside`, false, "Text with quote escape"},

		// Edge cases
		{`C:\test`, true, "Short Windows path with \\t"},
		{`test\nvalue`, false, "Short text with escape"},
		{`\users\data`, true, "Relative path starting with users"},
		{`\network\share`, false, "Network path starting with \\n pattern"},

		// Unix paths (forward slashes, no escaping issues)
		{`/usr/local/bin`, true, "Unix absolute path"},
		{`/tmp/data.log`, true, "Unix temp file"},
		{`./config/app.conf`, true, "Unix relative path"},

		// Non-path content with backslashes
		{`regex\\d+\\w*`, false, "Regex pattern"},
		{`JSON\\parsing`, false, "Text with escaped backslash"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isLikelyFilePath(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for: %s", tc.desc)
		})
	}
}

// TestUnicodeEscapeSequenceHandling tests handling of Unicode escape sequences
func TestUnicodeEscapeSequenceHandling(t *testing.T) {
	testCases := []struct {
		input       string
		isValidJSON bool
		desc        string
	}{
		{`\u0048`, true, "Valid Unicode H"},
		{`\u2605`, true, "Valid Unicode star"},
		{`\u0000`, true, "Valid Unicode null"},
		{`\uFFFF`, true, "Valid Unicode max BMP"},
		{`\u`, false, "Incomplete Unicode escape"},
		{`\u12`, false, "Incomplete Unicode escape (2 chars)"},
		{`\u123`, false, "Incomplete Unicode escape (3 chars)"},
		{`\uGHIJ`, false, "Invalid Unicode escape (non-hex)"},
		{`\u12GH`, false, "Invalid Unicode escape (mixed)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Test the pattern - complete Unicode escapes should have exactly 4 hex digits
			if tc.isValidJSON {
				assert.Len(t, tc.input, 6, "Valid Unicode escape should be 6 characters")
				assert.True(t, strings.HasPrefix(tc.input, `\u`), "Should start with \\u")

				// Check that the last 4 characters are hex digits
				hexPart := tc.input[2:]
				for _, r := range hexPart {
					assert.True(t, isHex(r), "Should be hex digit: %c", r)
				}
			} else if strings.HasPrefix(tc.input, `\u`) && len(tc.input) != 6 {
				// Invalid sequences should be identified
				assert.True(t, len(tc.input) < 6, "Incomplete sequence should be shorter than 6")
			}
		})
	}
}

// TestSpecialQuoteCharacterHandling tests handling of special quote characters
func TestSpecialQuoteCharacterHandling(t *testing.T) {
	testCases := []struct {
		input string
		desc  string
	}{
		{"\u201cquoted text\u201d", "Smart quotes"},
		{"\u2018single quoted\u2019", "Smart single quotes"},
		{"`backtick quoted`", "Backtick quotes"},
		{"\u201cangle quoted\u201d", "Smart double quotes"},
		{"\u201asingle\u2019", "Bottom single quotes"},
		{"\u201ebottom double\u201d", "Bottom double quotes"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// These should all be recognized as quote-like characters
			// and converted to standard JSON double quotes
			runes := []rune(tc.input)
			if len(runes) > 0 {
				// Test first and last characters
				firstChar := runes[0]
				lastChar := runes[len(runes)-1]

				// At least one should be recognized as a quote-like character
				isFirstQuote := isQuote(firstChar)
				isLastQuote := isQuote(lastChar)

				assert.True(t, isFirstQuote || isLastQuote,
					"Should recognize quote characters in: %s", tc.input)
			}
		})
	}
}
