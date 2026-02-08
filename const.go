package jsonrepair

// Character code constants used for JSON parsing.
const (
	codeBackslash               = 0x5c // "\"
	codeSlash                   = 0x2f // "/"
	codeAsterisk                = 0x2a // "*"
	codeOpeningBrace            = 0x7b // "{"
	codeClosingBrace            = 0x7d // "}"
	codeOpeningBracket          = 0x5b // "["
	codeClosingBracket          = 0x5d // "]"
	codeOpenParenthesis         = 0x28 // "("
	codeCloseParenthesis        = 0x29 // ")"
	codeSpace                   = 0x20 // " "
	codeNewline                 = 0xa  // "\n"
	codeTab                     = 0x9  // "\t"
	codeReturn                  = 0xd  // "\r"
	codeBackspace               = 0x08 // "\b"
	codeFormFeed                = 0x0c // "\f"
	codeDoubleQuote             = 0x22 // "
	codePlus                    = 0x2b // "+"
	codeMinus                   = 0x2d // "-"
	codeQuote                   = 0x27 // "'"
	codeZero                    = 0x30 // "0"
	codeNine                    = 0x39 // "9"
	codeComma                   = 0x2c // ","
	codeDot                     = 0x2e // "." (dot, period)
	codeColon                   = 0x3a // ":"
	codeSemicolon               = 0x3b // ";"
	codeUppercaseA              = 0x41 // "A"
	codeLowercaseA              = 0x61 // "a"
	codeUppercaseE              = 0x45 // "E"
	codeLowercaseE              = 0x65 // "e"
	codeUppercaseF              = 0x46 // "F"
	codeLowercaseF              = 0x66 // "f"
	codeNonBreakingSpace        = 0xa0   // non-breaking space
	codeEnQuad                  = 0x2000 // en quad (start of special whitespace range)
	codeHairSpace               = 0x200a // hair space (end of special whitespace range)
	codeNarrowNoBreakSpace      = 0x202f // narrow no-break space
	codeMediumMathematicalSpace = 0x205f // medium mathematical space
	codeIdeographicSpace        = 0x3000 // ideographic space
	codeDoubleQuoteLeft         = 0x201c // “
	codeDoubleQuoteRight        = 0x201d // ”
	codeQuoteLeft               = 0x2018 // ‘
	codeQuoteRight              = 0x2019 // ’
	codeGraveAccent             = 0x60   // `
	codeAcuteAccent             = 0xb4   // ´
)

// controlCharacters maps control character runes to their JSON escape sequences
// according to the JSON standard (RFC 8259).
var controlCharacters = map[rune]string{
	codeBackspace: `\b`,
	codeFormFeed:  `\f`,
	codeNewline:   `\n`,
	codeReturn:    `\r`,
	codeTab:       `\t`,
}

// escapeCharacters maps JSON standard escape characters to their unescaped values.
// These MUST be escaped or CAN be escaped in JSON strings.
var escapeCharacters = map[rune]string{
	'"':  "\"", // MUST be escaped
	'\\': "\\", // MUST be escaped
	'/':  "/",  // CAN be escaped (optional)
	'b':  "\b", // Backspace control character
	'f':  "\f", // Form feed control character
	'n':  "\n", // Newline control character
	'r':  "\r", // Carriage return control character
	't':  "\t", // Tab control character
	// Note: 'u' is handled separately for Unicode escape sequences (\uXXXX)
}
