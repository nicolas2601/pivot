package accounts

import "github.com/google/uuid"

// uuidParse is a small wrapper to keep handler.go tidy.
func uuidParse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}