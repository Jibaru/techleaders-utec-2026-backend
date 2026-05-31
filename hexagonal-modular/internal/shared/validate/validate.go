// Package validate holds format-level input validators used by the service
// layer. Every error returned here wraps ErrInvalidInput, so the controller's
// error mapper can translate any validation failure into a single 400 case.
//
// Validators expect already-trimmed input. Services trim before calling.
package validate

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// ErrInvalidInput is the wrapped sentinel for every validation failure.
// Callers use errors.Is(err, validate.ErrInvalidInput) to detect bad input.
var ErrInvalidInput = errors.New("invalid input")

const (
	minNameLen    = 2
	maxNameLen    = 100
	maxEmailLen   = 254 // RFC 5321
	maxAmountCent = 10_000_000
)

var externalIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{3,128}$`)

// Name checks length, allowed characters, and that there is at least one
// letter — so the value cannot be only digits or only punctuation.
func Name(name string) error {
	runes := []rune(name)
	if len(runes) < minNameLen || len(runes) > maxNameLen {
		return fmt.Errorf("%w: name must be %d-%d characters", ErrInvalidInput, minNameLen, maxNameLen)
	}
	hasLetter := false
	for _, r := range runes {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case r == ' ', r == '-', r == '\'', r == '.':
			// allowed separators inside a name
		default:
			return fmt.Errorf("%w: name contains invalid character %q", ErrInvalidInput, r)
		}
	}
	if !hasLetter {
		return fmt.Errorf("%w: name must contain at least one letter", ErrInvalidInput)
	}
	return nil
}

// Email parses with net/mail and rejects display-name forms.
func Email(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("%w: email is empty", ErrInvalidInput)
	}
	if len(email) > maxEmailLen {
		return fmt.Errorf("%w: email too long (max %d)", ErrInvalidInput, maxEmailLen)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: email is not valid: %v", ErrInvalidInput, err)
	}
	if parsed.Name != "" || !strings.EqualFold(parsed.Address, email) {
		return fmt.Errorf("%w: email must be a bare address", ErrInvalidInput)
	}
	return nil
}

func Amount(cents int64) error {
	if cents <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidInput)
	}
	if cents > maxAmountCent {
		return fmt.Errorf("%w: amount exceeds maximum of %d cents", ErrInvalidInput, maxAmountCent)
	}
	return nil
}

func ExternalID(id string) error {
	if !externalIDRe.MatchString(id) {
		return fmt.Errorf("%w: external_payment_id must be 3-128 chars of letters, digits, '.', '_', '-'", ErrInvalidInput)
	}
	return nil
}
