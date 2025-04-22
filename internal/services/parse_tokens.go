package services

import (
	"auth_service/internal/config"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// parseAccessToken разбирает access-токен и возвращает его содержимое в виде структуры TokenClaims.
func parseAccessToken(accessToken string) (*TokenClaims, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		secret := []byte(config.Secret)
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse jwt: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}

	payLoad, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to typecast to jwt.MapClaims")
	}

	userId, ok := payLoad["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'user_id' to string")
	}
	issuedIp, ok := payLoad["issued_ip"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'issued_ip' to string")
	}
	jti, ok := payLoad["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'jti' to string")
	}

	tokenClaims := &TokenClaims{
		UserId:   userId,
		IssuedIp: issuedIp,
		Jti:      jti,
	}

	return tokenClaims, nil
}

// parseRefreshToken разбирает refresh-токен, закодированный в формате base64, и возвращает его содержимое в виде структуры TokenClaims.
func parseRefreshToken(refreshToken string) (*TokenClaims, error) {
	var tokenClaims TokenClaims

	jsonBytes, err := base64.StdEncoding.DecodeString(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decoded base64 format: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &tokenClaims); err != nil {
		return nil, fmt.Errorf("failed to JSON decode: %w", err)
	}

	return &tokenClaims, nil
}

// checkRefreshToken проверяет хэш refresh-токена на соответствие с валидным хэшем.
func checkRefreshToken(refreshToken, validRefrTokenHash string) error {
	hash := sha512.Sum512([]byte(refreshToken))

	if err := bcrypt.CompareHashAndPassword([]byte(validRefrTokenHash), hash[:]); err != nil {
		return fmt.Errorf("refresh token hash is invalid: %w", err)
	}

	return nil
}
