package services

import "auth_service/internal/entities"

// AuthServiceInterface - интерфейс для работы с токенами аутентификации.
// Определяет методы для генерации и обновления токенов.
type AuthServiceInterface interface {
	GenerateTokens(userId, ip string) (*entities.TokensPair, error)
	RefreshTokens(ip string, tokensPair *entities.TokensPair) (*entities.TokensPair, error)
}
