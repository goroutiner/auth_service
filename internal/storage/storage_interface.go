package storage

import "auth_service/internal/entities"

// StorageInterface определяет универсальный интерфейс для работы с различными хранилищами данных (in-memory и postgres).
type StorageInterface interface {
	SaveRefreshTokenRecord(userId string, refreshTokenRecord *entities.RefreshTokenRecord) error              // Сохраняет хэш refresh-токена и claims пользователя.
	UpdateRefreshTokenRecord(oldJti, userId string, newRefreshTokenRecord *entities.RefreshTokenRecord) error // Обновляет refresh-токен по старому jti.
	GetRefreshTokenRecord(jti, userId string) (*entities.RefreshTokenRecord, error)                           // Возвращает record токена по jti и userId.
	GetUserEmail(userId string) (string, error)                                                               // GetUserEmail возвращает email пользователя по его userId.

}
