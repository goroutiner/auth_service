package services

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"crypto/sha512"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// parseAccessToken разбирает access-токен.
func parseAccessToken(accessToken string) (*entities.AccessTokenClaims, error) {
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

	if payLoad["jti"] == nil {
		return nil, fmt.Errorf("'jti' in claims is empty")
	}
	jti, ok := payLoad["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'jti' to string")
	}

	if payLoad["user_id"] == nil {
		return nil, fmt.Errorf("'user_id' in claims is empty")
	}
	userId, ok := payLoad["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'user_id' to string")
	}

	if payLoad["created_at"] == nil {
		return nil, fmt.Errorf("'created_at' in claims is empty")
	}
	createdAt, ok := payLoad["created_at"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'created_at' to float64")
	}

	if payLoad["expired_at"] == nil {
		return nil, fmt.Errorf("'expired_at' in claims is empty")
	}
	expiredAt, ok := payLoad["expired_at"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to typecast 'expired_at' to float64")
	}

	accessTokenClaims := &entities.AccessTokenClaims{
		Jti:       jti,
		UserId:    userId,
		CreatedAt: time.Unix(int64(createdAt), 0),
		ExpiredAt: time.Unix(int64(expiredAt), 0),
	}

	return accessTokenClaims, nil
}

// checkRefreshToken проверяет хэш refresh-токена на соответствие с валидным хэшем.
func checkRefreshToken(refreshToken, validRefrTokenHash string) error {
	hash := sha512.Sum512([]byte(refreshToken))
	if err := bcrypt.CompareHashAndPassword([]byte(validRefrTokenHash), hash[:]); err != nil {
		return fmt.Errorf("refresh token hash is invalid: %w", err)
	}

	return nil
}
