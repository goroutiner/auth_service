package storage

// StorageInterface определяет интерфейс для работы с различными хранилищами данных в "in-memory" и "postgres" режимах
type StorageInterface interface {
	SaveRefreshTokenHash(userId, refrTokenHash string) error
	GetRefreshTokenHash(userId string) (string, error)
	UpdateRefreshTokenHash(userId, newRefrTokenHash string) error
	GetUserEmail(userId string) (string, error)
}
