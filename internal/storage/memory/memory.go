package memory

import (
	"auth_service/internal/entities"
	"fmt"
	"sync"
)

// Memory представляет хранилище для работы с токенами в оперативной памяти.
type Memory struct {
	refreshTokens map[string]entities.Client // refreshTokens хранит соответствие user_id -> Client
	mu            sync.RWMutex               // mu обеспечивает потокобезопасность операций с хранилищем
}

// NewMemoryStore создает новое хранилище в оперативной памяти.
func NewMemoryStore() *Memory {
	return &Memory{
		refreshTokens: make(map[string]entities.Client),
	}
}

// SaveRefreshTokenHash сохраняет хэш refresh-токена для указанного пользователя.
// Если запись с данным user_id уже существует, возвращает ошибку.
func (m *Memory) SaveRefreshTokenHash(userId, refrTokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, has := m.refreshTokens[userId]
	if has {
		return fmt.Errorf("refresh token already exists")
	}

	mockEmail := "user@gmail.com"
	m.refreshTokens[userId] = entities.Client{
		Email:            mockEmail,
		RefreshTokenHash: refrTokenHash,
	}

	return nil
}

// GetRefreshTokenHash возвращает хэш refresh-токена для указанного пользователя.
// Если запись с данным user_id не найдена, возвращает ошибку.
func (m *Memory) GetRefreshTokenHash(userId string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, has := m.refreshTokens[userId]
	if !has {
		return "", fmt.Errorf("refresh token hash is not found")
	}

	return client.RefreshTokenHash, nil
}

// UpdateRefreshTokenHash обновляет хэш refresh-токена для указанного пользователя.
// Если запись с данным user_id не найдена, возвращает ошибку.
func (m *Memory) UpdateRefreshTokenHash(userId, newRefrTokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, has := m.refreshTokens[userId]; !has {
		return fmt.Errorf("refresh token hash is not found")
	}

	mockEmail := "user@gmail.com"
	m.refreshTokens[userId] = entities.Client{
		Email:            mockEmail,
		RefreshTokenHash: newRefrTokenHash,
	}

	return nil
}

// GetUserEmail возвращает email пользователя по его user_id.
// Если запись с данным user_id не найдена, возвращает ошибку.
func (m *Memory) GetUserEmail(userId string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, has := m.refreshTokens[userId]; !has {
		return "", fmt.Errorf("refresh token hash is not found")
	}

	client := m.refreshTokens[userId]

	return client.Email, nil
}
