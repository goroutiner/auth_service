package memory

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
)

// Memory реализует in-memory хранилище для refresh-токенов пользователей.
// Используется для тестирования или работы в режиме без постоянного хранилища.
type Memory struct {
	tokenRecords map[string][]*entities.RefreshTokenRecord // userTokens хранит список активных токенов пользователя по userId.
	mu           sync.RWMutex                              // mu обеспечивает потокобезопасность операций с хранилищем.
}

// NewMemoryStore создает и возвращает новое in-memory хранилище токенов.
func NewMemoryStore() *Memory {
	return &Memory{
		tokenRecords: make(map[string][]*entities.RefreshTokenRecord),
	}
}

// SaveRefreshTokenRecord сохраняет хэш refresh-токена для указанного пользователя.
// Если у пользователя уже максимальное количество токенов, удаляет самый старый токен.
// Возвращает ошибку, если токен с таким jti уже существует.
func (m *Memory) SaveRefreshTokenRecord(userId string, refreshTokenRecord *entities.RefreshTokenRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	maxTokensPerUser, err := strconv.Atoi(config.MaxTokensPerUser)
	if err != nil {
		return fmt.Errorf("env 'MAX_TOKENS_PER_USER' is not number: %w", err)
	}
	if err := m.checkActiveTokens(userId, maxTokensPerUser); err != nil {
		log.Println(err)
		if err := m.deleteOldestRefreshToken(userId); err != nil {
			return fmt.Errorf("failed to delete oldest refresh token for userID: '%s': %w", userId, err)
		}
		log.Printf("the latest token has been deleted due to exceeding the limit for userID: '%s'\n", userId)
	}
	_, has := m.tokenRecords[userId]
	if !has {
		m.tokenRecords[userId] = make([]*entities.RefreshTokenRecord, 0, maxTokensPerUser)
	}

	for _, record := range m.tokenRecords[userId] {
		if refreshTokenRecord.Jti == record.Jti {
			return fmt.Errorf("hash of refresh token already exists")
		}
	}
	m.tokenRecords[userId] = append(m.tokenRecords[userId], refreshTokenRecord)

	return nil
}

// UpdateRefreshTokenRecord обновляет refresh-токен пользователя по старому jti.
// Удаляет старый токен и добавляет новый. Возвращает ошибку, если токен не найден.
func (m *Memory) UpdateRefreshTokenRecord(oldJti, userId string, newRefreshTokenRecord *entities.RefreshTokenRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, has := m.tokenRecords[userId]
	if !has {
		return fmt.Errorf("user with userID: '%s' was not found", userId)
	}

	for i, record := range m.tokenRecords[userId] {
		if oldJti == record.Jti {
			m.tokenRecords[userId] = slices.Delete(m.tokenRecords[userId], i, i+1)
			m.tokenRecords[userId] = append(m.tokenRecords[userId], newRefreshTokenRecord)
			return nil
		}
	}

	return fmt.Errorf("hash of refresh token was not found")
}

// GetRefreshTokenRecord возвращает record токена пользователя по jti и userId.
// Если токен не найден, возвращает ошибку.
func (m *Memory) GetRefreshTokenRecord(jti, userId string) (*entities.RefreshTokenRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, has := m.tokenRecords[userId]
	if !has {
		return nil, fmt.Errorf("user with userID: '%s' was not found", userId)
	}
	for _, record := range m.tokenRecords[userId] {
		if jti == record.Jti {
			return record, nil
		}
	}

	return nil, fmt.Errorf("token record was not found")
}

// GetUserEmail возвращает email пользователя (в данном случае моковые данные).
// Если email не найден, возвращает ошибку.
func (d *Memory) GetUserEmail(userId string) (string, error) {
	mockEmail := "user@gmail.com"

	return mockEmail, nil
}

// checkActiveTokens проверяет количество активных refresh токенов для пользователя.
// Если лимит превышен, возвращает специальную ошибку.
func (m *Memory) checkActiveTokens(userId string, maxTokensPerUser int) error {
	if len(m.tokenRecords[userId]) == maxTokensPerUser {
		return fmt.Errorf("exceeding the limit for userID: '%s'", userId)
	}

	return nil
}

// deleteOldestRefreshToken удаляет самый старый refresh-токен пользователя.
// Если количество токенов не равно лимиту, возвращает ошибку.
func (m *Memory) deleteOldestRefreshToken(userId string) error {
	if len(m.tokenRecords[userId]) < 2 {
		return fmt.Errorf("number of user is less than max limit")
	}
	m.tokenRecords[userId] = m.tokenRecords[userId][1:]

	return nil
}
