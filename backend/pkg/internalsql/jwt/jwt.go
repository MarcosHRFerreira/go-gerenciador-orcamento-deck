package jwt

import (
	"errors"
	"strings"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

const accessTokenTTL = 60 * time.Minute

type Claims struct {
	UserID             int64  `json:"user_id"`
	Username           string `json:"username"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
	golangjwt.RegisteredClaims
}

func CreateToken(userID int64, username string, role string, mustChangePassword bool, secretKey string) (string, error) {
	claims := Claims{
		UserID:             userID,
		Username:           username,
		Role:               role,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: golangjwt.RegisteredClaims{
			ExpiresAt: golangjwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  golangjwt.NewNumericDate(time.Now()),
		},
	}

	token := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenStr string, secretKey string, validateClaims bool) (int64, string, string, bool, error) {
	claims := &Claims{}
	parserOptions := make([]golangjwt.ParserOption, 0, 1)
	if !validateClaims {
		parserOptions = append(parserOptions, golangjwt.WithoutClaimsValidation())
	}

	token, err := golangjwt.ParseWithClaims(extractToken(tokenStr), claims, func(token *golangjwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}, parserOptions...)
	if err != nil {
		return 0, "", "", false, err
	}

	if !token.Valid {
		return 0, "", "", false, errors.New("token invalido")
	}

	if claims.UserID == 0 {
		return 0, "", "", false, errors.New("user_id do token invalido")
	}
	if claims.Username == "" {
		return 0, "", "", false, errors.New("username do token invalido")
	}
	if claims.Role == "" {
		return 0, "", "", false, errors.New("role do token invalido")
	}

	return claims.UserID, claims.Username, claims.Role, claims.MustChangePassword, nil
}

func extractToken(headerValue string) string {
	const bearerPrefix = "Bearer "

	if strings.HasPrefix(headerValue, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(headerValue, bearerPrefix))
	}

	return strings.TrimSpace(headerValue)
}
