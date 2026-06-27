package middleware

import (
	"errors"
	"testing"
)

func TestExtractBearer(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantErr   error
	}{
		// Missing token cases — header empty or no token after prefix.
		{
			name:      "empty header",
			header:    "",
			wantToken: "",
			wantErr:   ErrMissingBearerToken,
		},
		{
			name:      "bearer with no token",
			header:    "Bearer ",
			wantToken: "",
			wantErr:   ErrMissingBearerToken,
		},
		{
			name:      "bearer with only whitespace",
			header:    "Bearer    ",
			wantToken: "",
			wantErr:   ErrMissingBearerToken,
		},

		// Malformed cases — wrong scheme or missing space.
		{
			name:      "bearer concatenated with token (no space)",
			header:    "Bearertoken123",
			wantToken: "",
			wantErr:   ErrMalformedAuthHeader,
		},
		{
			name:      "basic scheme",
			header:    "Basic dXNlcjpwYXNz",
			wantToken: "",
			wantErr:   ErrMalformedAuthHeader,
		},
		{
			name:      "no scheme at all",
			header:    "abc.def.ghi",
			wantToken: "",
			wantErr:   ErrMalformedAuthHeader,
		},
		{
			name:      "single token without scheme",
			header:    "nospacenosplit",
			wantToken: "",
			wantErr:   ErrMalformedAuthHeader,
		},

		// Valid cases — case-insensitive Bearer prefix, anything after the space.
		{
			name:      "standard bearer with token",
			header:    "Bearer abc.def.ghi",
			wantToken: "abc.def.ghi",
			wantErr:   nil,
		},
		{
			name:      "lowercase bearer prefix",
			header:    "bearer abc.def.ghi",
			wantToken: "abc.def.ghi",
			wantErr:   nil,
		},
		{
			name:      "uppercase bearer prefix",
			header:    "BEARER abc.def.ghi",
			wantToken: "abc.def.ghi",
			wantErr:   nil,
		},
		{
			name:      "mixed case bearer prefix",
			header:    "BeArEr abc.def.ghi",
			wantToken: "abc.def.ghi",
			wantErr:   nil,
		},
		{
			name:      "token with internal whitespace preserved",
			header:    "Bearer abc def ghi",
			wantToken: "abc def ghi",
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractBearer(tt.header)
			if token != tt.wantToken {
				t.Errorf("ExtractBearer(%q) token = %q, want %q", tt.header, token, tt.wantToken)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ExtractBearer(%q) err = %v, want %v", tt.header, err, tt.wantErr)
			}
		})
	}
}
