package httpfile

import (
	"fmt"
)

// ParseErrorType represents different types of parsing errors
type ParseErrorType int

const (
	ErrUnexpectedContent ParseErrorType = iota
	ErrMissingURL
	ErrInvalidURL
	ErrInvalidHeader
	ErrIncompleteRequest
	ErrMissingMethod
	ErrTemplateError
	ErrJSONError
	ErrMultilineHeader
)

// ParseError is a custom error type for HTTP file parsing errors
type ParseError struct {
	Type       ParseErrorType
	Message    string
	Line       string
	LineNumber int
	Err        error
}

func (e *ParseError) Error() string {
	if e.Line != "" {
		if e.LineNumber > 0 {
			return fmt.Sprintf("parse error: %s (line %d: %q)", e.Message, e.LineNumber, e.Line)
		}
		return fmt.Sprintf("parse error: %s (line: %q)", e.Message, e.Line)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func (e *ParseError) Is(target error) bool {
	if t, ok := target.(*ParseError); ok {
		return e.Type == t.Type
	}
	return false
}

// NewParseError creates a new ParseError
func NewParseError(errType ParseErrorType, message string, line string) *ParseError {
	return &ParseError{
		Type:    errType,
		Message: message,
		Line:    line,
	}
}

// NewParseErrorWithCause creates a new ParseError with an underlying cause
func NewParseErrorWithCause(errType ParseErrorType, message string, line string, err error) *ParseError {
	return &ParseError{
		Type:    errType,
		Message: message,
		Line:    line,
		Err:     err,
	}
}

// EnrichParseError enriches an existing ParseError with line context if it's missing
func EnrichParseError(err error, line string, lineNumber int) error {
	if parseErr, ok := err.(*ParseError); ok {
		if parseErr.Line == "" {
			parseErr.Line = line
		}
		if parseErr.LineNumber == 0 {
			parseErr.LineNumber = lineNumber
		}
	}
	return err
}
