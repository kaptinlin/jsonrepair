package jsonrepair

import "errors"

// Define error types for specific JSON repair issues
var (
	ErrUnexpectedEnd       = errors.New("unexpected end of json string")
	ErrObjectKeyExpected   = errors.New("object key expected")
	ErrColonExpected       = errors.New("colon expected")
	ErrInvalidCharacter    = errors.New("invalid character")
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrInvalidUnicode      = errors.New("invalid unicode character")
)
