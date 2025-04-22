package services

import (
	"auth_service/internal/config"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// GenAccessToken генерирует access token на основе данных из TokenClaims.
// Использует алгоритм SHA512 для подписи jwt токена.
func (tc *TokenClaims) GenAccessToken() (string, error) {
	tokenClaims := jwt.MapClaims{
		"user_id":   tc.UserId,
		"issued_ip": tc.IssuedIp,
		"jti":       tc.Jti,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, tokenClaims)

	accessToken, err := jwtToken.SignedString([]byte(config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token %w", err)
	}

	return accessToken, nil
}

// GenRefreshToken генерирует refresh token на основе данных из TokenClaims.
// Использует json сериализацию и представляет в формате base64.
func (tc *TokenClaims) GenRefreshToken() (string, error) {
	jsonBytes, err := json.Marshal(tc)
	if err != nil {
		return "", fmt.Errorf("failed to JSON encoded %w", err)
	}

	refreshToken := base64.StdEncoding.EncodeToString(jsonBytes)

	return refreshToken, nil
}

// GenJti генерирует уникальный идентификатор токена (JTI).
func GenJti() (string, error) {
	bytes := make([]byte, 128)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate jti: %w", err)
	}

	jti := base64.StdEncoding.EncodeToString(bytes)

	return jti, nil
}

// GenBcryptHash генерирует bcrypt-хэш для переданного refresh token.
// Использует для хеширования алгоритмы SHA-512 и bcrypt.
func GenBcryptHash(refreshToken string) (string, error) {
	hash := sha512.Sum512([]byte(refreshToken))

	refrTokenHash, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}

	return string(refrTokenHash), err
}
