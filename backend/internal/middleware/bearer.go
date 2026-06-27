package middleware

import (
	"errors"
	"strings"
)

// Sentinel errors returned by ExtractBearer. Callers can use errors.Is to
// distinguish missing vs malformed headers and return the appropriate 401 code.
var (
	ErrMissingBearerToken = errors.New("missing bearer token")
	ErrMalformedAuthHeader = errors.New("malformed authorization header")
)

// ExtractBearer parses an Authorization header value and returns the bearer
// token. The "Bearer" prefix is matched case-insensitively per RFC 6750 §2.1.
//
// Returns ErrMissingBearerToken when the header is empty or has no token after
// the "Bearer" prefix.
// Returns ErrMalformedAuthHeader when the header does not match the
// "Bearer <token>" shape (missing space, wrong scheme, etc.).
//
// Examples:
//
//	ExtractBearer("")                              → "", ErrMissingBearerToken
//	ExtractBearer("Bearer ")                       → "", ErrMissingBearerToken
//	ExtractBearer("Bearertoken123")                → "", ErrMalformedAuthHeader
//	ExtractBearer("Basic dXNlcjpwYXNz")            → "", ErrMalformedAuthHeader
//	ExtractBearer("Bearer abc.def.ghi")            → "abc.def.ghi", nil
//	ExtractBearer("bearer abc.def.ghi")            → "abc.def.ghi", nil
func ExtractBearer(header string) (string, error) {
	if header == "" {
		return "", ErrMissingBearerToken
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", ErrMalformedAuthHeader
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrMalformedAuthHeader
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrMissingBearerToken
	}
	return token, nil
}
