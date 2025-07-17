package services

import (
	"auth_service/internal/entities"
	"auth_service/internal/storage"
	"fmt"
	"log"
	"time"
)

// AuthService предоставляет методы для работы с токенами аутентификации пользователя.
// Включает генерацию, обновление и валидацию access/refresh токенов.
type AuthService struct {
	storage storage.StorageInterface // Интерфейс для взаимодействия с хранилищем данных (БД или память)
}

// NewAuthService создает новый экземпляр AuthService с указанным хранилищем.
func NewAuthService(s storage.StorageInterface) *AuthService {
	return &AuthService{storage: s}
}

// GenerateTokens генерирует новую пару токенов (access и refresh) для пользователя.
func (s *AuthService) GenerateTokens(userId, ip string) (*entities.TokensPair, error) {
	jti, err := GenJti()
	if err != nil {
		return nil, fmt.Errorf("failed to generate jti: %w", err)
	}

	accessTokenClaims := &entities.AccessTokenClaims{
		Jti:       jti,
		UserId:    userId,
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	accessToken, err := GenAccessToken(accessTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refrTokenHash, err := GenBcryptHash(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash: %w", err)
	}
	refreshTokenRecord := &entities.RefreshTokenRecord{
		Jti:       jti,
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(3 * 24 * time.Hour),
		IssuedIp:  ip,
		TokenHash: refrTokenHash,
	}
	if err := s.storage.SaveRefreshTokenRecord(userId, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	tokensPair := &entities.TokensPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	log.Printf("Access/refresh tokens issued for userID: '%s', jti: '%s', ip: '%s'\n", userId, jti, ip)

	return tokensPair, nil
}

// RefreshTokens обновляет пару токенов (access и refresh) для пользователя.
// Проверяет валидность старых токенов, валидирует refresh token, при необходимости отправляет уведомление о смене IP.
func (s *AuthService) RefreshTokens(ip string, tokensPair *entities.TokensPair) (*entities.TokensPair, error) {
	accessTokenClaims, err := parseAccessToken(tokensPair.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	refreshTokenRecord, err := s.storage.GetRefreshTokenRecord(accessTokenClaims.Jti, accessTokenClaims.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get token claims: %w", err)
	}
	if err := checkRefreshToken(tokensPair.RefreshToken, refreshTokenRecord.TokenHash); err != nil {
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}
	if refreshTokenRecord.IssuedIp != ip {
		if err = CheckConfigVar(); err != nil {
			return nil, fmt.Errorf("config variable is empty: %w", err)
		}
		mockUserEmail, err := s.storage.GetUserEmail(accessTokenClaims.UserId)
		if err != nil {
			return nil, fmt.Errorf("failed to get user email: %w", err)
		}
		if err = SendWarningMsg(mockUserEmail, refreshTokenRecord.IssuedIp, ip); err != nil {
			return nil, fmt.Errorf("failed to send warning message to user's Email: %w", err)
		}
	}

	newJti, err := GenJti()
	if err != nil {
		return nil, fmt.Errorf("failed to generate jti: %w", err)
	}

	newAccessTokenClaims := &entities.AccessTokenClaims{
		Jti:       newJti,
		UserId:    accessTokenClaims.UserId,
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	newAccessToken, err := GenAccessToken(newAccessTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := GenRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	newRefrTokenHash, err := GenBcryptHash(newRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash: %w", err)
	}
	newRefreshTokenRecord := &entities.RefreshTokenRecord{
		Jti:       newJti,
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(3 * 24 * time.Hour),
		IssuedIp:  ip,
		TokenHash: newRefrTokenHash,
	}
	if err = s.storage.UpdateRefreshTokenRecord(refreshTokenRecord.Jti, accessTokenClaims.UserId, newRefreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to update refresh token hash: %w", err)
	}

	newTokensPair := &entities.TokensPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	log.Printf("Access/refresh tokens refreshed for userID: '%s', old jti: '%s', new jti: '%s', ip: '%s'\n",
		accessTokenClaims.UserId, accessTokenClaims.Jti, newJti, ip)

	return newTokensPair, nil
}
