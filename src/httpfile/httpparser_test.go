package httpfile

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseSimpleGETRequest(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_get.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/simple_get.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) == 0 {
		t.Fatalf("No requests parsed")
	}
	if parser.reqs[0].Method != "GET" {
		t.Errorf("Expected method GET, got %s", parser.reqs[0].Method)
	}
	if parser.reqs[0].URL.String() != "http://example.com" {
		t.Errorf("Expected URL http://example.com, got %s", parser.reqs[0].URL.String())
	}
}

func TestParsePOSTWithHeadersAndBody(t *testing.T) {
	data, err := os.ReadFile("testdata/post_with_headers_and_body.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/post_with_headers_and_body.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) == 0 {
		t.Fatalf("No requests parsed")
	}
	parsedReq := parser.reqs[0]
	if parsedReq.Method != "POST" {
		t.Errorf("Expected method POST, got %s", parsedReq.Method)
	}
	if parsedReq.URL.String() != "http://example.com/api/resource" {
		t.Errorf("Expected URL http://example.com/api/resource, got %s", parsedReq.URL.String())
	}
	if parsedReq.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header, got %s", parsedReq.Header.Get("Content-Type"))
	}
	if parsedReq.Header.Get("Authorization") != "Bearer token123" {
		t.Errorf("Expected Authorization header, got %s", parsedReq.Header.Get("Authorization"))
	}
	bodyBytes, _ := io.ReadAll(parsedReq.Body)
	parsedReq.Body.Close()
	if strings.TrimSpace(string(bodyBytes)) != "{\"key\": \"value\"}" {
		t.Errorf("Expected body '{\"key\": \"value\"}', got '%s'", string(bodyBytes))
	}
}

func TestParseMultipleRequests(t *testing.T) {
	data, err := os.ReadFile("testdata/multiple_requests.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/multiple_requests.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(parser.reqs))
	}
	if parser.reqs[0].Method != "GET" || parser.reqs[0].URL.String() != "http://example.com/first" {
		t.Errorf("First request not parsed correctly: %+v", parser.reqs[0])
	}
	if parser.reqs[0].Header.Get("Header1") != "value1" {
		t.Errorf("First request headers not parsed correctly: %+v", parser.reqs[0].Header)
	}
	if parser.reqs[1].Method != "POST" || parser.reqs[1].URL.String() != "http://example.com/second" {
		t.Errorf("Second request not parsed correctly: %+v", parser.reqs[1])
	}
	if parser.reqs[1].Header.Get("Header2") != "value2" {
		t.Errorf("Second request headers not parsed correctly: %+v", parser.reqs[1].Header)
	}
	bodyBytes, _ := io.ReadAll(parser.reqs[1].Body)
	parser.reqs[1].Body.Close()
	if strings.TrimSpace(string(bodyBytes)) != "{\"foo\": \"bar\"}" {
		t.Errorf("Second request body not parsed correctly: '%s'", string(bodyBytes))
	}
}

func TestParseMalformedRequest(t *testing.T) {
	data, err := os.ReadFile("testdata/malformed_request.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/malformed_request.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err == nil {
		t.Fatalf("Expected parser error for malformed request, but got none")
	}
	// The first line "GET" (without URL) should trigger this specific error
	expectedError := "line does not match expected format"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error '%s', got: %v", expectedError, err)
	}
}

func TestParseEmptyFile(t *testing.T) {
	parser := newParser("")
	err := parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) != 0 {
		t.Errorf("Expected no requests for empty file, got %d", len(parser.reqs))
	}
}

func TestParseOnlySeparators(t *testing.T) {
	sepText := "###\n###\n"
	parser := newParser(sepText)
	err := parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) != 0 {
		t.Errorf("Expected no requests for only separators, got %d", len(parser.reqs))
	}
}

func TestParsePUTRequest(t *testing.T) {
	data, err := os.ReadFile("testdata/put_request.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/put_request.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) == 0 {
		t.Fatalf("No requests parsed")
	}
	if parser.reqs[0].Method != "PUT" {
		t.Errorf("Expected method PUT, got %s", parser.reqs[0].Method)
	}
	if parser.reqs[0].URL.String() != "http://example.com/put" {
		t.Errorf("Expected URL http://example.com/put, got %s", parser.reqs[0].URL.String())
	}
}

func TestParseDELETERequest(t *testing.T) {
	data, err := os.ReadFile("testdata/delete_request.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/delete_request.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) == 0 {
		t.Fatalf("No requests parsed")
	}
	if parser.reqs[0].Method != "DELETE" {
		t.Errorf("Expected method DELETE, got %s", parser.reqs[0].Method)
	}
	if parser.reqs[0].URL.String() != "http://example.com/delete" {
		t.Errorf("Expected URL http://example.com/delete, got %s", parser.reqs[0].URL.String())
	}
}

func TestParseNoHeadersNoBody(t *testing.T) {
	data, err := os.ReadFile("testdata/no_headers_no_body.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/no_headers_no_body.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}
	if len(parser.reqs) == 0 {
		t.Fatalf("No requests parsed")
	}
	if parser.reqs[0].Method != "GET" {
		t.Errorf("Expected method GET, got %s", parser.reqs[0].Method)
	}
	if parser.reqs[0].URL.String() != "http://example.com/noheaders" {
		t.Errorf("Expected URL http://example.com/noheaders, got %s", parser.reqs[0].URL.String())
	}
	if len(parser.reqs[0].Header) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(parser.reqs[0].Header))
	}
	bodyBytes, _ := io.ReadAll(parser.reqs[0].Body)
	parser.reqs[0].Body.Close()
	if strings.TrimSpace(string(bodyBytes)) != "" {
		t.Errorf("Expected empty body, got '%s'", string(bodyBytes))
	}
}

func TestParseMultilineHeader(t *testing.T) {
	data, err := os.ReadFile("testdata/multiline_header.http")
	if err != nil {
		t.Fatalf("Failed to read testdata/multiline_header.http: %v", err)
	}
	httpText := string(data)
	parser := newParser(httpText)
	err = parser.parse(false)
	if err == nil {
		t.Fatalf("Expected parser error for multiline header, but got none")
	}
	// Multiline headers should trigger this specific error
	expectedError := "headers must contain a colon separator"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error '%s', got: %v", expectedError, err)
	}
}

func TestParseErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		errorType ParseErrorType
		message   string
	}{
		{
			name:      "unexpected content",
			content:   "invalid line content\nGET http://example.com",
			errorType: ErrUnexpectedContent,
			message:   "line does not match expected format",
		},
		{
			name:      "missing URL",
			content:   "GET \n",
			errorType: ErrMissingURL,
			message:   "HTTP method must be followed by a URL",
		},
		{
			name:      "invalid header",
			content:   "GET http://example.com\nInvalid Header Without Colon",
			errorType: ErrInvalidHeader,
			message:   "headers must contain a colon separator",
		},
		{
			name:      "invalid URL - no scheme",
			content:   "GET example.com",
			errorType: ErrInvalidURL,
			message:   "URL must have a scheme",
		},
		{
			name:      "invalid URL - no host",
			content:   "GET http://",
			errorType: ErrInvalidURL,
			message:   "URL must have a host",
		},
		{
			name:      "invalid URL - malformed",
			content:   "GET http://[invalid-host",
			errorType: ErrInvalidURL,
			message:   "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newParser(tt.content)
			err := parser.parse(false)

			if err == nil {
				t.Fatalf("Expected error but got none")
			}

			// Test errors.As functionality
			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Expected ParseError, got %T", err)
			}

			if parseErr.Type != tt.errorType {
				t.Errorf("Expected error type %v, got %v", tt.errorType, parseErr.Type)
			}

			if !strings.Contains(parseErr.Message, tt.message) {
				t.Errorf("Expected message to contain %q, got %q", tt.message, parseErr.Message)
			}

			// Test errors.Is functionality
			targetErr := NewParseError(tt.errorType, "", "")
			if !errors.Is(err, targetErr) {
				t.Errorf("errors.Is should return true for same error type")
			}

			// Test that different error types return false
			differentErr := NewParseError(ErrTemplateError, "", "")
			if errors.Is(err, differentErr) {
				t.Errorf("errors.Is should return false for different error type")
			}
		})
	}
}

func TestParseErrorWithCause(t *testing.T) {
	originalErr := errors.New("original error")
	parseErr := NewParseErrorWithCause(ErrTemplateError, "template failed", "some line", originalErr)

	// Test that Unwrap works correctly
	if errors.Unwrap(parseErr) != originalErr {
		t.Errorf("Expected unwrapped error to be original error")
	}

	// Test that errors.Is works with wrapped errors
	if !errors.Is(parseErr, originalErr) {
		t.Errorf("errors.Is should find the wrapped error")
	}
}

func TestParseErrorStringRepresentation(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name:     "with line",
			err:      NewParseError(ErrUnexpectedContent, "test message", "test line"),
			expected: `parse error: test message (line: "test line")`,
		},
		{
			name:     "without line",
			err:      NewParseError(ErrUnexpectedContent, "test message", ""),
			expected: "parse error: test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		shouldErr bool
		errorType ParseErrorType
	}{
		{
			name:      "valid HTTP URL",
			url:       "http://example.com",
			shouldErr: false,
		},
		{
			name:      "valid HTTPS URL",
			url:       "https://example.com/path",
			shouldErr: false,
		},
		{
			name:      "empty URL",
			url:       "",
			shouldErr: true,
			errorType: ErrMissingURL,
		},
		{
			name:      "no scheme",
			url:       "example.com",
			shouldErr: true,
			errorType: ErrInvalidURL,
		},
		{
			name:      "no host",
			url:       "http://",
			shouldErr: true,
			errorType: ErrInvalidURL,
		},
		{
			name:      "malformed URL",
			url:       "http://[invalid",
			shouldErr: true,
			errorType: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}

				var parseErr *ParseError
				if !errors.As(err, &parseErr) {
					t.Fatalf("Expected ParseError, got %T", err)
				}

				if parseErr.Type != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, parseErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
