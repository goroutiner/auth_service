package services

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// GenAccessToken генерирует access token (JWT) на основе данных из AccessTokenClaims.
// В payload токена включаются user_id, issued_ip и jti.
// Для подписи используется алгоритм SHA512 (HS512) и секрет из конфига.
func GenAccessToken(accessTokenClaims *entities.AccessTokenClaims) (string, error) {
	tokenClaims := jwt.MapClaims{
		"jti":        accessTokenClaims.Jti,
		"user_id":    accessTokenClaims.UserId,
		"created_at": accessTokenClaims.CreatedAt.Unix(),
		"expired_at": accessTokenClaims.ExpiredAt.Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, tokenClaims)

	accessToken, err := jwtToken.SignedString([]byte(config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token %w", err)
	}

	return accessToken, nil
}

// GenRefreshToken генерирует криптографически стойкий refresh token.
// Использует 32 случайных байта, кодирует их в base64.
func GenRefreshToken() (string, error) {
	src := make([]byte, 32)
	_, err := rand.Read(src)
	if err != nil {
		return "", fmt.Errorf("failed to generate random source for refresh token %w", err)
	}
	refreshToken := base64.StdEncoding.EncodeToString(src)

	return refreshToken, nil
}

// GenJti генерирует уникальный идентификатор токена (JTI).
// Использует 32 случайных байта, кодирует их в base64.
func GenJti() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate jti: %w", err)
	}
	jti := base64.StdEncoding.EncodeToString(bytes)

	return jti, nil
}

// GenBcryptHash генерирует bcrypt-хэш для переданного refresh token.
// Сначала вычисляет SHA-512 от токена, затем хеширует результат с помощью bcrypt.
func GenBcryptHash(refreshToken string) (string, error) {
	hash := sha512.Sum512([]byte(refreshToken))
	refrTokenHash, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}

	return string(refrTokenHash), err
}
