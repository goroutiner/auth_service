package services

import (
	"auth_service/internal/entities"
	"auth_service/internal/storage"
	"fmt"
)

// TokenClaims представляет структуру для хранения данных токенов.
type TokenClaims struct {
	UserId   string // Идентификатор пользователя
	IssuedIp string // IP-адрес, с которого был выдан токен
	Jti      string // Уникальный идентификатор токена
}

// AuthService предоставляет методы для работы с токенами аутентификации.
type AuthService struct {
	storage storage.StorageInterface // Интерфейс для взаимодействия с хранилищем данных
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(s storage.StorageInterface) *AuthService {
	return &AuthService{storage: s}
}

// GenerateTokens генерирует пару токенов (доступа и обновления) для указанного пользователя.
func (s *AuthService) GenerateTokens(userId, ip string) (*entities.TokensPair, error) {
	jti, err := GenJti()
	if err != nil {
		return nil, fmt.Errorf("failed to generate jti: %w", err)
	}

	tokenClaims := &TokenClaims{
		UserId:   userId,
		IssuedIp: ip,
		Jti:      jti,
	}

	accessToken, err := tokenClaims.GenAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := tokenClaims.GenRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refrTokenHash, err := GenBcryptHash(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash: %w", err)
	}

	if err := s.storage.SaveRefreshTokenHash(userId, refrTokenHash); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	tokensPair := &entities.TokensPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return tokensPair, nil
}

// RefreshTokens обновляет пару токенов (доступа и обновления).
func (s *AuthService) RefreshTokens(ip string, tokensPair *entities.TokensPair) (*entities.TokensPair, error) {
	accessTokenClaims, err := parseAccessToken(tokensPair.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	refreshTokenClaims, err := parseRefreshToken(tokensPair.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if *accessTokenClaims != *refreshTokenClaims {
		return nil, fmt.Errorf("pair of tokens was not issued together")
	}

	refreshToken, err := refreshTokenClaims.GenRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	userId := refreshTokenClaims.UserId

	validRefrTokenHash, err := s.storage.GetRefreshTokenHash(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token hash: %w", err)
	}

	if err := checkRefreshToken(refreshToken, validRefrTokenHash); err != nil {
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}

	issuedIp := refreshTokenClaims.IssuedIp
	if issuedIp != ip {
		userEmail, err := s.storage.GetUserEmail(userId)
		if err != nil {
			return nil, fmt.Errorf("failed to get user Email: %w", err)
		}

		if err = checkConfigVar(); err != nil {
			return nil, fmt.Errorf("config variable is empty: %w", err)
		}

		if err = sendWarningMsg(userEmail, issuedIp, ip); err != nil {
			return nil, fmt.Errorf("failed to send warning message to user's Email: %w", err)
		}
	}

	newJti, err := GenJti()
	if err != nil {
		return nil, fmt.Errorf("failed to generate jti: %w", err)
	}

	newTokenClaims := &TokenClaims{
		UserId:   userId,
		IssuedIp: ip,
		Jti:      newJti,
	}

	newAccessToken, err := newTokenClaims.GenAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := newTokenClaims.GenRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newRefrTokenHash, err := GenBcryptHash(newRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash: %w", err)
	}

	if err = s.storage.UpdateRefreshTokenHash(userId, string(newRefrTokenHash)); err != nil {
		return nil, fmt.Errorf("failed to update refresh token hash")
	}

	newTokensPair := &entities.TokensPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	return newTokensPair, nil
}
