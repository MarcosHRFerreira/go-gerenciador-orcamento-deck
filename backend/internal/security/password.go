package security

import (
	"strings"
	"unicode"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

const strongPasswordMessage = "password must contain at least 8 characters, uppercase letter, lowercase letter, number and special character"

func ValidateStrongPassword(password string) error {
	passwordLength := len(password)
	if passwordLength < minPasswordLength || passwordLength > maxPasswordLength {
		return apperror.BadRequest(strongPasswordMessage)
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, character := range password {
		switch {
		case unicode.IsUpper(character):
			hasUpper = true
		case unicode.IsLower(character):
			hasLower = true
		case unicode.IsDigit(character):
			hasNumber = true
		case strings.ContainsRune(`!"#$%&'()*+,-./:;<=>?@[\]^_{|}~`, character):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return apperror.BadRequest(strongPasswordMessage)
	}

	return nil
}
