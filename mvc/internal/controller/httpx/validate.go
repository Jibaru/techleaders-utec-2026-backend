package httpx

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidName       = errors.New("name is not valid")
	ErrInvalidEmail      = errors.New("email is not valid")
	ErrInvalidAmount     = errors.New("amount is not valid")
	ErrInvalidExternalID = errors.New("external id is not valid")
)

const (
	minNameLen    = 2
	maxNameLen    = 100
	maxEmailLen   = 254 // RFC 5321
	maxAmountCent = 10_000_000
)

var externalIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{3,128}$`)

// ValidateName checks the length, allowed characters, and that there is at
// least one letter — so the value cannot be only digits or only punctuation.
// Hyphens, apostrophes, and periods are allowed for names like "Anne-Marie",
// "O'Brien", or "Dr.".  Uses Unicode letter classification so non-ASCII
// names work without ceremony.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	runes := []rune(name)
	if len(runes) < minNameLen || len(runes) > maxNameLen {
		return fmt.Errorf("%w: must be %d-%d characters", ErrInvalidName, minNameLen, maxNameLen)
	}
	hasLetter := false
	for _, r := range runes {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case r == ' ', r == '-', r == '\'', r == '.':
			// allowed separators inside a name
		default:
			return fmt.Errorf("%w: contains invalid character %q", ErrInvalidName, r)
		}
	}
	if !hasLetter {
		return fmt.Errorf("%w: must contain at least one letter", ErrInvalidName)
	}
	return nil
}

// ValidateEmail parses the address with net/mail and rejects display-name
// forms like "John <jd@example.com>" so only bare addresses pass.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) == 0 {
		return fmt.Errorf("%w: empty", ErrInvalidEmail)
	}
	if len(email) > maxEmailLen {
		return fmt.Errorf("%w: too long (max %d)", ErrInvalidEmail, maxEmailLen)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	}
	if parsed.Name != "" || !strings.EqualFold(parsed.Address, email) {
		return fmt.Errorf("%w: must be a bare address", ErrInvalidEmail)
	}
	return nil
}

// ValidateAmount rejects non-positive amounts and absurdly large values.
// The ceiling is a sanity check, not a business rule.
func ValidateAmount(cents int64) error {
	if cents <= 0 {
		return fmt.Errorf("%w: must be positive", ErrInvalidAmount)
	}
	if cents > maxAmountCent {
		return fmt.Errorf("%w: exceeds maximum of %d cents", ErrInvalidAmount, maxAmountCent)
	}
	return nil
}

// ValidateExternalID checks the format of an external payment processor id:
// 3-128 chars, letters / digits / dot / underscore / dash only. Whitespace and
// control characters are rejected.
func ValidateExternalID(id string) error {
	if !externalIDRe.MatchString(id) {
		return fmt.Errorf("%w: must be 3-128 chars of letters, digits, '.', '_', '-'", ErrInvalidExternalID)
	}
	return nil
}
