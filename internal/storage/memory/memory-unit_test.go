package memory_test

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"auth_service/internal/storage/memory"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var store *memory.Memory

// TestMain инициализирует in-memory хранилище и устанавливает параметры конфигурации для тестов.
func TestMain(m *testing.M) {
	config.MaxTokensPerUser = "5"
	store = memory.NewMemoryStore()
	code := m.Run()
	os.Exit(code)
}

// TestSaveRefreshTokenRecord проверяет сохранение refresh токена в памяти, включая ошибочные кейсы.
func TestSaveRefreshTokenRecord(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	now := time.Now().UTC().Truncate(time.Microsecond)
	refreshTokenRecord := &entities.RefreshTokenRecord{
		Jti:       "jti123",
		CreatedAt: now,
		ExpiredAt: now.Add(24 * time.Hour),
		IssuedIp:  "192.168.0.1",
		TokenHash: "hash123",
	}
	err := store.SaveRefreshTokenRecord(userId, refreshTokenRecord)
	require.NoError(t, err)

	actual, err := store.GetRefreshTokenRecord(refreshTokenRecord.Jti, userId)
	require.NoError(t, err)
	require.Equal(t, refreshTokenRecord, actual)

	t.Run("duplicate add", func(t *testing.T) {
		err := store.SaveRefreshTokenRecord(userId, refreshTokenRecord)
		require.Error(t, err)
		require.ErrorContains(t, err, "already exists")
	})
}

// TestUpdateRefreshTokenRecord проверяет обновление refresh токена в памяти, включая ошибочные кейсы.
func TestUpdateRefreshTokenRecord(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	now := time.Now().UTC().Truncate(time.Microsecond)
	refreshTokenRecord := &entities.RefreshTokenRecord{
		Jti:       "jti123",
		CreatedAt: now,
		ExpiredAt: now.Add(24 * time.Hour),
		IssuedIp:  "192.168.0.1",
		TokenHash: "hash123",
	}
	err := store.SaveRefreshTokenRecord(userId, refreshTokenRecord)
	require.NoError(t, err)

	newNow := now.Add(1 * time.Hour)
	newRefreshTokenRecord := &entities.RefreshTokenRecord{
		Jti:       "jti456",
		CreatedAt: newNow,
		ExpiredAt: newNow.Add(48 * time.Hour),
		IssuedIp:  "192.168.0.2",
		TokenHash: "hash456",
	}
	err = store.UpdateRefreshTokenRecord(refreshTokenRecord.Jti, userId, newRefreshTokenRecord)
	require.NoError(t, err)

	actual, err := store.GetRefreshTokenRecord(newRefreshTokenRecord.Jti, userId)
	require.NoError(t, err)
	require.Equal(t, newRefreshTokenRecord, actual)

	t.Run("update non-existent", func(t *testing.T) {
		err := store.UpdateRefreshTokenRecord("not_exist_jti", userId, newRefreshTokenRecord)
		require.Error(t, err)
		require.ErrorContains(t, err, "not found")
	})
}

// TestGetTokenRecord проверяет получение refresh токена из памяти, включая ошибочные кейсы.
func TestGetTokenRecord(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	now := time.Now().UTC().Truncate(time.Microsecond)
	refreshTokenRecord := &entities.RefreshTokenRecord{
		Jti:       "jti123",
		CreatedAt: now,
		ExpiredAt: now.Add(24 * time.Hour),
		IssuedIp:  "192.168.0.1",
		TokenHash: "hash123",
	}
	err := store.SaveRefreshTokenRecord(userId, refreshTokenRecord)
	require.NoError(t, err)

	actual, err := store.GetRefreshTokenRecord(refreshTokenRecord.Jti, userId)
	require.NoError(t, err)
	require.Equal(t, refreshTokenRecord, actual)

	t.Run("get non-existent", func(t *testing.T) {
		_, err := store.GetRefreshTokenRecord("not_exist_jti", userId)
		require.Error(t, err)
		require.ErrorContains(t, err, "not found")
	})
}
