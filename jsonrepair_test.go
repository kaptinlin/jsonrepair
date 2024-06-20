package jsonrepair

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFullJSONObject tests parsing a full JSON object.
func TestParseFullJSONObject(t *testing.T) {
	text := `{"a":2.3e100,"b":"str","c":null,"d":false,"e":[1,2,3]}`
	parsed, err := JSONRepair(text)
	require.NoError(t, err)
	assert.Equal(t, text, parsed)
}

// TestParseWhitespace tests parsing JSON with whitespace.
func TestParseWhitespace(t *testing.T) {
	assertRepairEqual(t, "  { \n } \t ")
}

// TestParseObject tests parsing JSON objects.
func TestParseObject(t *testing.T) {
	assertRepairEqual(t, "{}")
	assertRepairEqual(t, "{  }")
	assertRepairEqual(t, `{"a": {}}`)
	assertRepairEqual(t, `{"a": "b"}`)
	assertRepairEqual(t, `{"a": 2}`)
}

// TestParseArray tests parsing JSON arrays.
func TestParseArray(t *testing.T) {
	assertRepairEqual(t, "[]")
	assertRepairEqual(t, "[  ]")
	assertRepairEqual(t, "[1,2,3]")
	assertRepairEqual(t, "[ 1 , 2 , 3 ]")
	assertRepairEqual(t, "[1,2,[3,4,5]]")
	assertRepairEqual(t, "[{}]")
	assertRepairEqual(t, `{"a":[]}`)
	assertRepairEqual(t, `[1, "hi", true, false, null, {}, []]`)
}

// TestParseNumber tests parsing JSON numbers.
func TestParseNumber(t *testing.T) {
	assertRepairEqual(t, "23")
	assertRepairEqual(t, "0")
	assertRepairEqual(t, "0e+2")
	assertRepairEqual(t, "0.0")
	assertRepairEqual(t, "-0")
	assertRepairEqual(t, "2.3")
	assertRepairEqual(t, "2300e3")
	assertRepairEqual(t, "2300e+3")
	assertRepairEqual(t, "2300e-3")
	assertRepairEqual(t, "-2")
	assertRepairEqual(t, "2e-3")
	assertRepairEqual(t, "2.3e-3")
}

// TestParseString tests parsing JSON strings.
func TestParseString(t *testing.T) {
	assertRepairEqual(t, `"str"`)
	assertRepairEqual(t, "\"\\\"\\\\\\/\\b\\f\\n\\r\\t\"")
	assertRepairEqual(t, `"\\u260E"`)
}

// TestParseKeywords tests parsing JSON keywords.
func TestParseKeywords(t *testing.T) {
	assertRepairEqual(t, "true")
	assertRepairEqual(t, "false")
	assertRepairEqual(t, "null")
}

// TestCorrectlyHandleStringsEqualingDelimiter tests handling strings that equal a JSON delimiter.
func TestCorrectlyHandleStringsEqualingDelimiter(t *testing.T) {
	assertRepairEqual(t, `""`)
	assertRepairEqual(t, `"["`)
	assertRepairEqual(t, `"]"`)
	assertRepairEqual(t, `"{"`)
	assertRepairEqual(t, `"}"`)
	assertRepairEqual(t, `":"`)
	assertRepairEqual(t, `","`)
}

// TestSupportsUnicodeCharactersInString tests parsing strings with Unicode characters.
func TestSupportsUnicodeCharactersInString(t *testing.T) {
	assertRepairEqual(t, `"★"`)
	assertRepairEqual(t, `"\u2605"`)
	assertRepairEqual(t, `"😀"`)
	assertRepairEqual(t, `"\ud83d\ude00"`)
	assertRepairEqual(t, `"йнформация"`)
}

// TestSupportsEscapedUnicodeCharactersInString tests parsing strings with escaped Unicode characters.
func TestSupportsEscapedUnicodeCharactersInString(t *testing.T) {
	assertRepairEqual(t, `"\\u2605"`)
	assertRepairEqual(t, `"\\u2605A"`)
	assertRepairEqual(t, `"\\ud83d\\ude00"`)
	assertRepairEqual(t, `"\\u0439\\u043d\\u0444\\u043e\\u0440\\u043c\\u0430\\u0446\\u0438\\u044f"`)
}

// TestSupportsUnicodeCharactersInKey tests parsing JSON objects with Unicode characters in keys.
func TestSupportsUnicodeCharactersInKey(t *testing.T) {
	assertRepairEqual(t, `{"★":true}`)
	assertRepairEqual(t, `{"\u2605":true}`)
	assertRepairEqual(t, `{"😀":true}`)
	assertRepairEqual(t, `{"\ud83d\ude00":true}`)
}

// TestShouldAddMissingQuotes tests repairing missing quotes in JSON.
func TestShouldAddMissingQuotes(t *testing.T) {
	assertRepair(t, `abc`, `"abc"`)
	assertRepair(t, `hello   world`, `"hello   world"`)
	assertRepair(t, "{\nmessage: hello world\n}", "{\n\"message\": \"hello world\"\n}")
	assertRepair(t, `{a:2}`, `{"a":2}`)
	assertRepair(t, `{a: 2}`, `{"a": 2}`)
	assertRepair(t, `{2: 2}`, `{"2": 2}`)
	assertRepair(t, `{true: 2}`, `{"true": 2}`)
	assertRepair(t, "{\n  a: 2\n}", "{\n  \"a\": 2\n}")
	assertRepair(t, `[a,b]`, `["a","b"]`)
	assertRepair(t, "[\na,\nb\n]", "[\n\"a\",\n\"b\"\n]")
}

// TestShouldAddMissingEndQuote tests repairing missing end quotes in JSON.
func TestShouldAddMissingEndQuote(t *testing.T) {
	assertRepair(t, `"abc`, `"abc"`)
	assertRepair(t, `'abc`, `"abc"`)
	assertRepair(t, "\u2018abc", `"abc"`)
	assertRepair(t, `"it's working`, `"it's working"`)
	assertRepair(t, `["abc+/*comment*/"def"]`, `["abcdef"]`)
	assertRepair(t, `["abc/*comment*/+"def"]`, `["abcdef"]`)
	assertRepair(t, `["abc,/*comment*/"def"]`, `["abc","def"]`)
}

// TestShouldRepairTruncatedJSON tests repairing truncated JSON.
func TestShouldRepairTruncatedJSON(t *testing.T) {
	assertRepair(t, `"foo`, `"foo"`)
	assertRepair(t, `[`, `[]`)
	assertRepair(t, `["foo`, `["foo"]`)
	assertRepair(t, `["foo"`, `["foo"]`)
	assertRepair(t, `["foo",`, `["foo"]`)
	assertRepair(t, `{"foo":"bar"`, `{"foo":"bar"}`)
	assertRepair(t, `{"foo":"bar`, `{"foo":"bar"}`)
	assertRepair(t, `{"foo":`, `{"foo":null}`)
	assertRepair(t, `{"foo"`, `{"foo":null}`)
	assertRepair(t, `{"foo`, `{"foo":null}`)
	assertRepair(t, `{`, `{}`)
	assertRepair(t, `2.`, `2.0`)
	assertRepair(t, `2e`, `2e0`)
	assertRepair(t, `2e+`, `2e+0`)
	assertRepair(t, `2e-`, `2e-0`)
	assertRepair(t, `{"foo":"bar\u20`, `{"foo":"bar"}`)
	assertRepair(t, `"\u`, `""`)
	assertRepair(t, `"\u2`, `""`)
	assertRepair(t, `"\u260`, `""`)
	assertRepair(t, `"\u2605`, `"\u2605"`)
	assertRepair(t, `{"s \ud`, `{"s": null}`)
	assertRepair(t, `{"message": "it's working`, `{"message": "it's working"}`)
	assertRepair(t, `{"text":"Hello Sergey,I hop`, `{"text":"Hello Sergey,I hop"}`)
	assertRepair(t, `{"message": "with, multiple, commma's, you see?`, `{"message": "with, multiple, commma's, you see?"}`)
}

// TestShouldRepairEllipsisInArray tests repairing ellipses in JSON arrays.
func TestShouldRepairEllipsisInArray(t *testing.T) {
	assertRepair(t, `[1,2,3,...]`, `[1,2,3]`)
	assertRepair(t, `[1, 2, 3, ... ]`, `[1, 2, 3  ]`)
	assertRepair(t, `[1,2,3,/*comment1*/.../*comment2*/]`, `[1,2,3]`)
	assertRepair(t, "[\n  1,\n  2,\n  3,\n  /*comment1*/  .../*comment2*/\n]", "[\n  1,\n  2,\n  3\n    \n]")
	assertRepair(t, `{"array":[1,2,3,...]}`, `{"array":[1,2,3]}`)
	assertRepair(t, `[1,2,3,...,9]`, `[1,2,3,9]`)
	assertRepair(t, `[...,7,8,9]`, `[7,8,9]`)
	assertRepair(t, `[..., 7,8,9]`, `[ 7,8,9]`)
	assertRepair(t, `[...]`, `[]`)
	assertRepair(t, `[ ... ]`, `[  ]`)
}

// TestShouldRepairEllipsisInObject tests repairing ellipses in JSON objects.
func TestShouldRepairEllipsisInObject(t *testing.T) {
	assertRepair(t, `{"a":2,"b":3,...}`, `{"a":2,"b":3}`)
	assertRepair(t, `{"a":2,"b":3,/*comment1*/.../*comment2*/}`, `{"a":2,"b":3}`)
	assertRepair(t, "{\n  \"a\":2,\n  \"b\":3,\n  /*comment1*/.../*comment2*/\n}", "{\n  \"a\":2,\n  \"b\":3\n  \n}")
	assertRepair(t, `{"a":2,"b":3, ... }`, `{"a":2,"b":3  }`)
	assertRepair(t, `{"nested":{"a":2,"b":3, ... }}`, `{"nested":{"a":2,"b":3  }}`)
	assertRepair(t, `{"a":2,"b":3,...,"z":26}`, `{"a":2,"b":3,"z":26}`)
	assertRepair(t, `{"a":2,"b":3,...}`, `{"a":2,"b":3}`)
	assertRepair(t, `{...}`, `{}`)
	assertRepair(t, `{ ... }`, `{  }`)
}

// TestShouldAddMissingStartQuote tests repairing missing start quotes in JSON.
func TestShouldAddMissingStartQuote(t *testing.T) {
	assertRepair(t, `abc"`, `"abc"`)
	assertRepair(t, `[a","b"]`, `["a","b"]`)
	assertRepair(t, `[a",b"]`, `["a","b"]`)
	assertRepair(t, `{"a":"foo","b":"bar"}`, `{"a":"foo","b":"bar"}`)
	assertRepair(t, `{a":"foo","b":"bar"}`, `{"a":"foo","b":"bar"}`)
	assertRepair(t, `{"a":"foo",b":"bar"}`, `{"a":"foo","b":"bar"}`)
	assertRepair(t, `{"a":foo","b":"bar"}`, `{"a":"foo","b":"bar"}`)
}

// TestShouldStopAtFirstNextReturnWhenMissingEndQuote tests stopping at the next return when missing an end quote.
func TestShouldStopAtFirstNextReturnWhenMissingEndQuote(t *testing.T) {
	assertRepair(t, "[\n\"abc,\n\"def\"\n]", "[\n\"abc\",\n\"def\"\n]")
	assertRepair(t, "[\n\"abc,  \n\"def\"\n]", "[\n\"abc\",  \n\"def\"\n]")
	assertRepair(t, "[\"abc]\n", "[\"abc\"]\n")
	assertRepair(t, "[\"abc  ]\n", "[\"abc\"  ]\n")
	assertRepair(t, "[\n[\n\"abc\n]\n]\n", "[\n[\n\"abc\"\n]\n]\n")
}

// TestShouldReplaceSingleQuotesWithDoubleQuotes tests replacing single quotes with double quotes in JSON.
func TestShouldReplaceSingleQuotesWithDoubleQuotes(t *testing.T) {
	assertRepair(t, "{'a':2}", "{\"a\":2}")
	assertRepair(t, "{'a':'foo'}", "{\"a\":\"foo\"}")
	assertRepair(t, "{\"a\":'foo'}", "{\"a\":\"foo\"}")
	assertRepair(t, "{a:'foo',b:'bar'}", "{\"a\":\"foo\",\"b\":\"bar\"}")
}

// TestShouldReplaceSpecialQuotesWithDoubleQuotes tests replacing special quotes with double quotes in JSON.
func TestShouldReplaceSpecialQuotesWithDoubleQuotes(t *testing.T) {
	assertRepair(t, "{“a”:“b”}", "{\"a\":\"b\"}")
	assertRepair(t, "{‘a’:‘b’}", "{\"a\":\"b\"}")
	assertRepair(t, "{`a´:`b´}", "{\"a\":\"b\"}")
}

// TestShouldNotReplaceSpecialQuotesInsideNormalString tests not replacing special quotes inside a normal string.
func TestShouldNotReplaceSpecialQuotesInsideNormalString(t *testing.T) {
	assertRepair(t, "\"Rounded “ quote\"", "\"Rounded “ quote\"")
	assertRepair(t, "'Rounded “ quote'", "\"Rounded “ quote\"")
	assertRepair(t, "\"Rounded ’ quote\"", "\"Rounded ’ quote\"")
	assertRepair(t, "'Rounded ’ quote'", "\"Rounded ’ quote\"")
	assertRepair(t, "'Double \\\" quote'", "\"Double \\\" quote\"")
}

// TestShouldNotCrashWhenRepairingQuotes tests not crashing when repairing quotes in JSON.
func TestShouldNotCrashWhenRepairingQuotes(t *testing.T) {
	assertRepair(t, "{pattern: '’'}", "{\"pattern\": \"’\"}")
}

// TestShouldLeaveStringContentUntouched tests leaving string content untouched in JSON.
func TestShouldLeaveStringContentUntouched(t *testing.T) {
	assertRepairEqual(t, `"{a:b}"`)
}

// TestShouldAddRemoveEscapeCharacters tests adding and removing escape characters in JSON strings.
func TestShouldAddRemoveEscapeCharacters(t *testing.T) {
	assertRepair(t, `"foo'bar"`, `"foo'bar"`)
	assertRepair(t, `"foo\"bar"`, `"foo\"bar"`)
	assertRepair(t, `'foo"bar'`, `"foo\"bar"`)
	assertRepair(t, `'foo\'bar'`, `"foo'bar"`)
	assertRepair(t, `"foo\'bar"`, `"foo'bar"`)
	assertRepair(t, `"\a"`, `"a"`)
}

// TestShouldRepairMissingObjectValue tests repairing missing object values in JSON.
func TestShouldRepairMissingObjectValue(t *testing.T) {
	assertRepair(t, `{"a":}`, `{"a":null}`)
	assertRepair(t, `{"a":,"b":2}`, `{"a":null,"b":2}`)
	assertRepair(t, `{"a":`, `{"a":null}`)
}

// TestShouldRepairUndefinedValues tests repairing undefined values in JSON.
func TestShouldRepairUndefinedValues(t *testing.T) {
	assertRepair(t, `{"a":undefined}`, `{"a":null}`)
	assertRepair(t, `[undefined]`, `[null]`)
	assertRepair(t, `undefined`, `null`)
}

// TestShouldEscapeUnescapedControlCharacters tests escaping unescaped control characters in JSON strings.
func TestShouldEscapeUnescapedControlCharacters(t *testing.T) {
	assertRepair(t, "\"hello\bworld\"", `"hello\bworld"`)
	assertRepair(t, "\"hello\fworld\"", `"hello\fworld"`)
	assertRepair(t, "\"hello\nworld\"", `"hello\nworld"`)
	assertRepair(t, "\"hello\rworld\"", `"hello\rworld"`)
	assertRepair(t, "\"hello\tworld\"", `"hello\tworld"`)
	assertRepair(t, "{\"key\nafter\": \"foo\"}", `{"key\nafter": "foo"}`)
	assertRepair(t, "[\"hello\nworld\"]", `["hello\nworld"]`)
	assertRepair(t, "[\"hello\nworld\"  ]", `["hello\nworld"  ]`)
	assertRepair(t, "[\"hello\nworld\"\n]", "[\"hello\\nworld\"\n]")
}

// TestShouldEscapeUnescapedDoubleQuotes tests escaping unescaped double quotes in JSON strings.
func TestShouldEscapeUnescapedDoubleQuotes(t *testing.T) {
	assertRepair(t, `"The TV has a 24" screen"`, `"The TV has a 24\" screen"`)
	assertRepair(t, `{"key": "apple "bee" carrot"}`, `{"key": "apple \"bee\" carrot"}`)
	assertRepairEqual(t, `[",",":"]`)
	assertRepair(t, `["a" 2]`, `["a", 2]`)
	assertRepair(t, `["a" 2`, `["a", 2]`)
	assertRepair(t, `["," 2`, `[",", 2]`)
}

// TestShouldReplaceSpecialWhiteSpaceCharacters tests replacing special white space characters in JSON strings.
func TestShouldReplaceSpecialWhiteSpaceCharacters(t *testing.T) {
	assertRepair(t, "{\"a\":\u00a0\"foo\u00a0bar\"}", "{\"a\": \"foo\u00a0bar\"}")
	assertRepair(t, "{\"a\":\u202F\"foo\"}", `{"a": "foo"}`)
	assertRepair(t, "{\"a\":\u205F\"foo\"}", `{"a": "foo"}`)
	assertRepair(t, "{\"a\":\u3000\"foo\"}", `{"a": "foo"}`)
}

// TestShouldReplaceNonNormalizedLeftRightQuotes tests replacing non-normalized left/right quotes in JSON strings.
func TestShouldReplaceNonNormalizedLeftRightQuotes(t *testing.T) {
	assertRepair(t, "\u2018foo\u2019", `"foo"`)
	assertRepair(t, "\u201Cfoo\u201D", `"foo"`)
	assertRepair(t, "\u0060foo\u00B4", `"foo"`)
	assertRepair(t, "\u0060foo'", `"foo"`)
	assertRepair(t, "\u0060foo'", `"foo"`)
}

// TestShouldRemoveBlockComments tests removing block comments from JSON strings.
func TestShouldRemoveBlockComments(t *testing.T) {
	assertRepair(t, "/* foo */ {}", " {}")
	assertRepair(t, "{} /* foo */ ", "{}  ")
	assertRepair(t, "{} /* foo ", "{} ")
	assertRepair(t, "\n/* foo */\n{}", "\n\n{}")
	assertRepair(t, `{"a":"foo",/*hello*/"b":"bar"}`, `{"a":"foo","b":"bar"}`)
	assertRepair(t, `{"flag":/*boolean*/true}`, `{"flag":true}`)
}

// TestShouldRemoveLineComments tests removing line comments in JSON.
func TestShouldRemoveLineComments(t *testing.T) {
	assertRepair(t, "{} // comment", "{} ")
	assertRepair(t, "{\n\"a\":\"foo\",//hello\n\"b\":\"bar\"\n}", "{\n\"a\":\"foo\",\n\"b\":\"bar\"\n}")
}

// TestShouldNotRemoveCommentsInsideString tests not removing comments inside a string in JSON.
func TestShouldNotRemoveCommentsInsideString(t *testing.T) {
	assertRepairEqual(t, `"/* foo */"`)
}

// TestShouldRemoveCommentsAfterStringContainingDelimiter tests removing comments after a string containing a delimiter.
func TestShouldRemoveCommentsAfterStringContainingDelimiter(t *testing.T) {
	assertRepair(t, `["a"/* foo */]`, `["a"]`)
	assertRepair(t, `["(a)"/* foo */]`, `["(a)"]`)
	assertRepair(t, `["a]"/* foo */]`, `["a]"]`)
	assertRepair(t, `{"a":"b"/* foo */}`, `{"a":"b"}`)
	assertRepair(t, `{"a":"(b)"/* foo */}`, `{"a":"(b)"}`)
}

// TestShouldStripJSONPNotation tests stripping JSONP notation in JSON.
func TestShouldStripJSONPNotation(t *testing.T) {
	// matching
	assertRepair(t, "callback_123({});", "{}")
	assertRepair(t, "callback_123([]);", "[]")
	assertRepair(t, "callback_123(2);", "2")
	assertRepair(t, `callback_123("foo");`, `"foo"`)
	assertRepair(t, "callback_123(null);", "null")
	assertRepair(t, "callback_123(true);", "true")
	assertRepair(t, "callback_123(false);", "false")
	assertRepair(t, "callback({})", "{}")
	assertRepair(t, "/* foo bar */ callback_123 ({})", " {}")
	assertRepair(t, "/* foo bar */ callback_123 ({})", " {}")
	assertRepair(t, "/* foo bar */\ncallback_123({})", "\n{}")
	assertRepair(t, "/* foo bar */ callback_123 (  {}  )", "   {}  ")
	assertRepair(t, "  /* foo bar */   callback_123({});  ", "     {}  ")
	assertRepair(t, "\n/* foo\nbar */\ncallback_123 ({});\n\n", "\n\n{}\n\n")
	// non-matching
	assertRepairFailure(t, `callback {}`, `unexpected character: '{'`, 9)
}

// TestShouldRepairEscapedStringContents tests repairing escaped string contents in JSON strings.
func TestShouldRepairEscapedStringContents(t *testing.T) {
	assertRepair(t, `\"hello world\"`, `"hello world"`)
	assertRepair(t, `\"hello world\`, `"hello world"`)
	assertRepair(t, `\"hello \\"world\\"\"`, `"hello \"world\""`)
	assertRepair(t, `[\"hello \\"world\\"\"]`, `["hello \"world\""]`)
	assertRepair(t, `{\"stringified\": \"hello \\"world\\"\"}`, `{"stringified": "hello \"world\""}`)

	// the following is a bit weird but comes close to the most likely intention
	// assertRepair(t, `[\"hello\, \"world\"]`, `["hello", "world"]`)

	// the following is sort of invalid: the end quote should be escaped too,
	// but the fixed result is most likely what you want in the end
	assertRepair(t, `\"hello"`, `"hello"`)
}

// TestShouldStripLeadingCommaFromArray tests stripping a leading comma from JSON arrays.
func TestShouldStripLeadingCommaFromArray(t *testing.T) {
	assertRepair(t, `[1,2,3]`, `[1,2,3]`)
	assertRepair(t, `[/* a */,/* b */1,2,3]`, `[1,2,3]`)
	assertRepair(t, `[ , 1,2,3]`, `[  1,2,3]`)
	assertRepair(t, `[ , 1,2,3]`, `[  1,2,3]`)
}

// TestShouldStripLeadingCommaFromObject tests stripping a leading comma from an object in JSON strings.
func TestShouldStripLeadingCommaFromObject(t *testing.T) {
	assertRepair(t, `{,"message": "hi"}`, `{"message": "hi"}`)
	assertRepair(t, `{/* a */,/* b */"message": "hi"}`, `{"message": "hi"}`)
	assertRepair(t, `{ ,"message": "hi"}`, `{ "message": "hi"}`)
	assertRepair(t, `{, "message": "hi"}`, `{ "message": "hi"}`)
}

// TestShouldStripTrailingCommasFromArray tests stripping trailing commas from JSON arrays.
func TestShouldStripTrailingCommasFromArray(t *testing.T) {
	assertRepair(t, "[1,2,3,]", "[1,2,3]")
	assertRepair(t, "[1,2,3,\n]", "[1,2,3\n]")
	assertRepair(t, "[1,2,3,  \n  ]", "[1,2,3  \n  ]")
	assertRepair(t, "[1,2,3,/*foo*/]", "[1,2,3]")
	assertRepair(t, "{\"array\":[1,2,3,]}", "{\"array\":[1,2,3]}")
	// not matching: inside a string
	assertRepair(t, "\"[1,2,3,]\"", "\"[1,2,3,]\"")
}

// TestShouldStripTrailingCommasFromObject tests stripping trailing commas from JSON objects.
func TestShouldStripTrailingCommasFromObject(t *testing.T) {
	assertRepair(t, "{\"a\":2,}", "{\"a\":2}")
	assertRepair(t, "{\"a\":2  ,  }", "{\"a\":2    }")
	assertRepair(t, "{\"a\":2  , \n }", "{\"a\":2   \n }")
	assertRepair(t, "{\"a\":2/*foo*/,/*foo*/}", "{\"a\":2}")
	assertRepair(t, "{},", "{}")
	// not matching: inside a string
	assertRepair(t, "\"{a:2,}\"", "\"{a:2,}\"")
}

// TestShouldStripTrailingCommaAtEnd tests stripping a trailing comma at the end of JSON.
func TestShouldStripTrailingCommaAtEnd(t *testing.T) {
	assertRepair(t, "4,", "4")
	assertRepair(t, "4 ,", "4 ")
	assertRepair(t, "4 , ", "4  ")
	assertRepair(t, "{\"a\":2},", "{\"a\":2}")
	assertRepair(t, "[1,2,3],", "[1,2,3]")
}

// TestShouldAddMissingClosingBraceForObject tests adding a missing closing brace for JSON objects.
func TestShouldAddMissingClosingBraceForObject(t *testing.T) {
	assertRepair(t, "{", "{}")
	assertRepair(t, "{\"a\":2", "{\"a\":2}")
	assertRepair(t, "{\"a\":2,", "{\"a\":2}")
	assertRepair(t, "{\"a\":{\"b\":2}", "{\"a\":{\"b\":2}}")
	assertRepair(t, "{\n  \"a\":{\"b\":2\n}", "{\n  \"a\":{\"b\":2\n}}")
	assertRepair(t, "[{\"b\":2]", "[{\"b\":2}]")
	assertRepair(t, "[{\"b\":2\n]", "[{\"b\":2}\n]")
	assertRepair(t, "[{\"i\":1{\"i\":2}]", "[{\"i\":1},{\"i\":2}]")
	assertRepair(t, "[{\"i\":1,{\"i\":2}]", "[{\"i\":1},{\"i\":2}]")
}

// TestShouldRemoveRedundantClosingBracketForObject tests removing a redundant closing bracket for JSON objects.
func TestShouldRemoveRedundantClosingBracketForObject(t *testing.T) {
	assertRepair(t, `{"a": 1}}`, `{"a": 1}`)
	assertRepair(t, `{"a": 1}}]}`, `{"a": 1}`)
	assertRepair(t, `{"a": 1 }  }  ]  }  `, `{"a": 1 }        `)
	assertRepair(t, `{"a":2]`, `{"a":2}`)
	assertRepair(t, `{"a":2,]`, `{"a":2}`)
	assertRepair(t, `{}}`, `{}`)
	assertRepair(t, `[2,}`, `[2]`)
	assertRepair(t, `[}`, `[]`)
	assertRepair(t, `{]`, `{}`)
}

// TestShouldAddMissingClosingBracketForArray tests adding a missing closing bracket for an array in JSON strings.
func TestShouldAddMissingClosingBracketForArray(t *testing.T) {
	assertRepair(t, "[", "[]")
	assertRepair(t, "[1,2,3", "[1,2,3]")
	assertRepair(t, "[1,2,3,", "[1,2,3]")
	assertRepair(t, "[[1,2,3,", "[[1,2,3]]")
	assertRepair(t, "{\n\"values\":[1,2,3\n}", "{\n\"values\":[1,2,3]\n}")
	assertRepair(t, "{\n\"values\":[1,2,3\n", "{\n\"values\":[1,2,3]}\n")
}

// TestShouldStripMongoDBDataTypes tests stripping MongoDB data types in JSON.
func TestShouldStripMongoDBDataTypes(t *testing.T) {
	// simple
	assertRepair(t, `NumberLong("2")`, `"2"`)
	assertRepair(t, `{"_id":ObjectId("123")}`, `{"_id":"123"}`)
	// extensive
	mongoDocument := `
		{
			"_id" : ObjectId("123"),
			"isoDate" : ISODate("2012-12-19T06:01:17.171Z"),
			"regularNumber" : 67,
			"long" : NumberLong("2"),
			"long2" : NumberLong(2),
			"int" : NumberInt("3"),
			"int2" : NumberInt(3),
			"decimal" : NumberDecimal("4"),
			"decimal2" : NumberDecimal(4)
		}`
	expectedJson := `
		{
			"_id" : "123",
			"isoDate" : "2012-12-19T06:01:17.171Z",
			"regularNumber" : 67,
			"long" : "2",
			"long2" : 2,
			"int" : "3",
			"int2" : 3,
			"decimal" : "4",
			"decimal2" : 4
		}`
	assertRepair(t, mongoDocument, expectedJson)
}

// TestShouldNotMatchMongoDBLikeFunctionsInUnquotedString tests not matching MongoDB-like functions in an unquoted string.
func TestShouldNotMatchMongoDBLikeFunctionsInUnquotedString(t *testing.T) {
	assertRepairFailure(t, `["This is C(2)", "This is F(3)]`, `unexpected character: '('`, 27)
	assertRepairFailure(t, `["This is C(2)", This is F(3)]`, `unexpected character: '('`, 26)
}

// TestShouldReplacePythonConstants tests replacing Python constants (None, True, False) in JSON.
func TestShouldReplacePythonConstants(t *testing.T) {
	assertRepair(t, `True`, `true`)
	assertRepair(t, `False`, `false`)
	assertRepair(t, `None`, `null`)
}

// TestShouldTurnUnknownSymbolsIntoString tests turning unknown symbols into a string in JSON strings.
func TestShouldTurnUnknownSymbolsIntoString(t *testing.T) {
	assertRepair(t, "foo", `"foo"`)
	assertRepair(t, "[1,foo,4]", `[1,"foo",4]`)
	assertRepair(t, "{foo: bar}", `{"foo": "bar"}`)

	assertRepair(t, "foo 2 bar", `"foo 2 bar"`)
	assertRepair(t, "{greeting: hello world}", `{"greeting": "hello world"}`)
	assertRepair(t, "{greeting: hello world\nnext: \"line\"}", "{\"greeting\": \"hello world\",\n\"next\": \"line\"}")
	assertRepair(t, "{greeting: hello world!}", `{"greeting": "hello world!"}`)
}

// TestShouldTurnInvalidNumbersIntoStrings tests turning invalid numbers into strings in JSON.
func TestShouldTurnInvalidNumbersIntoStrings(t *testing.T) {
	assertRepair(t, `ES2020`, `"ES2020"`)
	assertRepair(t, `0.0.1`, `"0.0.1"`)
	assertRepair(t, `746de9ad-d4ff-4c66-97d7-00a92ad46967`, `"746de9ad-d4ff-4c66-97d7-00a92ad46967"`)
	assertRepair(t, `234..5`, `"234..5"`)
	assertRepair(t, `[0.0.1,2]`, `["0.0.1",2]`)      // test delimiter for numerics
	assertRepair(t, `[2 0.0.1 2]`, `[2, "0.0.1 2"]`) // note: currently spaces delimit numbers, but don't delimit unquoted strings
	assertRepair(t, `2e3.4`, `"2e3.4"`)
}

// TestShouldRepairRegularExpressions tests repairing regular expressions in JSON.
func TestShouldRepairRegularExpressions(t *testing.T) {
	assertRepair(t, `{regex: /standalone-styles.css/}`, `{"regex": "/standalone-styles.css/"}`)
}

// TestShouldConcatenateStrings tests concatenating strings in JSON strings.
func TestShouldConcatenateStrings(t *testing.T) {
	assertRepair(t, `"hello" + " world"`, `"hello world"`)
	assertRepair(t, "\"hello\" +\n \" world\"", `"hello world"`)
	assertRepair(t, `"a"+"b"+"c"`, `"abc"`)
	assertRepair(t, `"hello" + /*comment*/ " world"`, `"hello world"`)
	assertRepair(t, "{\n  \"greeting\": 'hello' +\n 'world'\n}", "{\n  \"greeting\": \"helloworld\"\n}")

	assertRepair(t, "\"hello +\n \" world\"", `"hello world"`)
	assertRepair(t, `"hello +`, `"hello"`)
	assertRepair(t, `["hello +]`, `["hello"]`)
}

// TestShouldRepairMissingCommaBetweenArrayItems tests repairing missing comma between array items in JSON strings.
func TestShouldRepairMissingCommaBetweenArrayItems(t *testing.T) {
	assertRepair(t, `{"array": [{}{}]}`, `{"array": [{},{}]}`)
	assertRepair(t, `{"array": [{} {}]}`, `{"array": [{}, {}]}`)
	assertRepair(t, "{\"array\": [{}\n{}]}", "{\"array\": [{},\n{}]}")
	assertRepair(t, "{\"array\": [\n{}\n{}\n]}", "{\"array\": [\n{},\n{}\n]}")
	assertRepair(t, "{\"array\": [\n1\n2\n]}", "{\"array\": [\n1,\n2\n]}")
	assertRepair(t, "{\"array\": [\n\"a\"\n\"b\"\n]}", "{\"array\": [\n\"a\",\n\"b\"\n]}")

	// // should leave normal array as is
	assertRepair(t, "[\n{},\n{}\n]", "[\n{},\n{}\n]")
}

// TestShouldRepairMissingCommaBetweenObjectProperties tests repairing missing comma between object properties in JSON strings.
func TestShouldRepairMissingCommaBetweenObjectProperties(t *testing.T) {
	assertRepair(t, "{\"a\":2\n\"b\":3\n}", "{\"a\":2,\n\"b\":3\n}")
	assertRepair(t, "{\"a\":2\n\"b\":3\nc:4}", "{\"a\":2,\n\"b\":3,\n\"c\":4}")
}

// TestShouldRepairNumbersAtEnd tests repairing numbers at the end in JSON strings.
func TestShouldRepairNumbersAtEnd(t *testing.T) {
	assertRepair(t, `{"a":2.`, `{"a":2.0}`)
	assertRepair(t, `{"a":2e`, `{"a":2e0}`)
	assertRepair(t, `{"a":2e-`, `{"a":2e-0}`)
	assertRepair(t, `{"a":-`, `{"a":-0}`)
	assertRepair(t, `[2e,`, `[2e0]`)
	assertRepair(t, `[2e `, `[2e0] `) // spaces delimit numbers
	assertRepair(t, `[-,`, `[-0]`)
}

// TestShouldRepairMissingColonBetweenObjectKeyAndValue tests repairing missing colon between object key and value in JSON strings.
func TestShouldRepairMissingColonBetweenObjectKeyAndValue(t *testing.T) {
	assertRepair(t, `{"a" "b"}`, `{"a": "b"}`)
	assertRepair(t, `{"a" 2}`, `{"a": 2}`)
	assertRepair(t, `{"a" true}`, `{"a": true}`)
	assertRepair(t, `{"a" false}`, `{"a": false}`)
	assertRepair(t, `{"a" null}`, `{"a": null}`)
	assertRepair(t, `{"a"2}`, `{"a":2}`)
	assertRepair(t, "{\n\"a\" \"b\"\n}", "{\n\"a\": \"b\"\n}")
	assertRepair(t, `{"a" 'b'}`, `{"a": "b"}`)
	assertRepair(t, `{'a' 'b'}`, `{"a": "b"}`)
	assertRepair(t, `{“a” “b”}`, `{"a": "b"}`)
	assertRepair(t, `{a 'b'}`, `{"a": "b"}`)
	assertRepair(t, `{a “b”}`, `{"a": "b"}`)
}

// TestShouldRepairMissingCombinationOfCommaQuotesAndBrackets tests repairing missing combinations of comma, quotes, and brackets in JSON strings.
func TestShouldRepairMissingCombinationOfCommaQuotesAndBrackets(t *testing.T) {
	assertRepair(t, "{\"array\": [\na\nb\n]}", "{\"array\": [\n\"a\",\n\"b\"\n]}")
	assertRepair(t, "1\n2", "[\n1,\n2\n]")
	assertRepair(t, "[a,b\nc]", "[\"a\",\"b\",\n\"c\"]")
}

// TestShouldRepairNewlineSeparatedJSON tests repairing newline separated JSON (for example from MongoDB).
func TestShouldRepairNewlineSeparatedJSON(t *testing.T) {
	text := "/* 1 */\n{}\n\n/* 2 */\n{}\n\n/* 3 */\n{}\n"
	expected := "[\n\n{},\n\n\n{},\n\n\n{}\n\n]"
	assertRepair(t, text, expected)
}

// TestShouldRepairNewlineSeparatedJSONHavingCommas tests repairing newline separated JSON having commas.
func TestShouldRepairNewlineSeparatedJSONHavingCommas(t *testing.T) {
	text := "" + "/* 1 */\n" + "{},\n" + "\n" + "/* 2 */\n" + "{},\n" + "\n" + "/* 3 */\n" + "{}\n"
	expected := "[\n\n{},\n\n\n{},\n\n\n{}\n\n]"
	assertRepair(t, text, expected)
}

// TestShouldRepairNewlineSeparatedJSONHavingCommasAndTrailingComma tests repairing newline separated JSON having commas and trailing comma.
func TestShouldRepairNewlineSeparatedJSONHavingCommasAndTrailingComma(t *testing.T) {
	text := "" +
		"/* 1 */\n" +
		"{},\n" +
		"\n" +
		"/* 2 */\n" +
		"{},\n" +
		"\n" +
		"/* 3 */\n" +
		"{},\n"
	expected := "[\n\n{},\n\n\n{},\n\n\n{}\n\n]"

	assertRepair(t, text, expected)
}

// TestShouldRepairCommaSeparatedListWithValue tests repairing a comma-separated list with values.
func TestShouldRepairCommaSeparatedListWithValue(t *testing.T) {
	assertRepair(t, "1,2,3", "[\n1,2,3\n]")
	assertRepair(t, "1,2,3,", "[\n1,2,3\n]")
	assertRepair(t, "1\n2\n3", "[\n1,\n2,\n3\n]")
	assertRepair(t, "a\nb", "[\n\"a\",\n\"b\"\n]")
	assertRepair(t, "a,b", "[\n\"a\",\"b\"\n]")
}

// TestShouldRepairNumberWithLeadingZero tests repairing numbers with leading zeros in JSON strings.
func TestShouldRepairNumberWithLeadingZero(t *testing.T) {
	assertRepair(t, "0789", "\"0789\"")
	assertRepair(t, "000789", "\"000789\"")
	assertRepair(t, "001.2", "\"001.2\"")
	assertRepair(t, "002e3", "\"002e3\"")
	assertRepair(t, "[0789]", "[\"0789\"]")
	assertRepair(t, "{value:0789}", "{\"value\":\"0789\"}")
}

// TestShouldThrowExceptionInCaseOfNonRepairableIssues tests that the JSON repair throws an exception for non-repairable issues.
func TestShouldThrowExceptionInCaseOfNonRepairableIssues(t *testing.T) {
	assertRepairFailure(t, "", "unexpected end of json string", 0)
	// assertRepairFailure(t, `{"a",`, "colon expected", 4)
	// assertRepairFailure(t, "{:2}", "object key expected", 1)
	assertRepairFailure(t, `{"a":2}{}`, `unexpected character: '{'`, 7)
	// assertRepairFailure(t, `{"a" ]`, "colon expected", 5)
	assertRepairFailure(t, `{"a":2}foo`, `unexpected character: 'f'`, 7)
	assertRepairFailure(t, `foo [`, `unexpected character: '['`, 4)
	assertRepairEqual(t, `"\\u26"`)
	// assertRepairFailure(t, `"\\u26"`, `invalid unicode character '\\u26'`, 1)
	assertRepairEqual(t, `"\\uZ000"`)
	// assertRepairFailure(t, `"\\uZ000"`, `invalid unicode character '\\uZ000'`, 1)
	assertRepairEqual(t, `"\\uZ000"`)
	// assertRepairFailure(t, `"\\uZ000`, `invalid unicode character '\\uZ000'`, 1)
}

// assertRepairFailure is a helper function to check the JSON repair failure.
func assertRepairFailure(t *testing.T, text, expectedErrMsg string, expectedPos int) {
	result, err := JSONRepair(text)
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrMsg)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", expectedPos))
	assert.Empty(t, result)
}

func assertRepairEqual(t *testing.T, text string) {
	result, err := JSONRepair(text)
	require.NoError(t, err)
	assert.Equal(t, text, result)
}

func assertRepair(t *testing.T, text string, expected string) {
	result, err := JSONRepair(text)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
