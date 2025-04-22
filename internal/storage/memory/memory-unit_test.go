package memory_test

import (
	"auth_service/internal/storage/memory"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSaveRefreshTokenHash проверяет сохранение хэша refresh токена в памяти.
func TestSaveRefreshTokenHash(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	refreshTokenHash := "hash123"

	t.Run("successful saving token", func(t *testing.T) {
		err := store.SaveRefreshTokenHash(userId, refreshTokenHash)
		require.NoError(t, err)
	})

	t.Run("token already exists", func(t *testing.T) {
		err := store.SaveRefreshTokenHash(userId, "newHash456")
		require.Error(t, err)
		require.ErrorContains(t, err, "refresh token already exists")
	})
}

// TestGetRefreshTokenHash проверяет получение хэша refresh токена из памяти.
func TestGetRefreshTokenHash(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	refreshTokenHash := "hash123"

	err := store.SaveRefreshTokenHash(userId, refreshTokenHash)
	require.NoError(t, err)

	t.Run("successful getting token", func(t *testing.T) {
		actualRefreshTokenHash, err := store.GetRefreshTokenHash(userId)
		require.NoError(t, err)
		require.Equal(t, refreshTokenHash, actualRefreshTokenHash)
	})

	t.Run("invalid userId", func(t *testing.T) {
		invalidUserId := "user1234"
		_, err := store.GetRefreshTokenHash(invalidUserId)
		require.Error(t, err)
		require.ErrorContains(t, err, "refresh token hash is not found")
	})
}

// TestUpdateRefreshTokenHash проверяет обновление хэша refresh токена в памяти.
func TestUpdateRefreshTokenHash(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	oldRefreshTokenHash := "oldHash123"
	newRefreshTokenHash := "newHash456"

	err := store.SaveRefreshTokenHash(userId, oldRefreshTokenHash)
	require.NoError(t, err)

	t.Run("successful updating token", func(t *testing.T) {
		err = store.UpdateRefreshTokenHash(userId, newRefreshTokenHash)
		require.NoError(t, err)

		actualRefreshTokenHash, err := store.GetRefreshTokenHash(userId)
		require.NoError(t, err)
		require.Equal(t, newRefreshTokenHash, actualRefreshTokenHash)
	})

	t.Run("invalid userId", func(t *testing.T) {
		invalidUserId := "user1234"
		err := store.UpdateRefreshTokenHash(invalidUserId, newRefreshTokenHash)
		require.Error(t, err)
		require.ErrorContains(t, err, "refresh token hash is not found")
	})
}

// TestGetUserEmail проверяет получение email пользователя из памяти.
func TestGetUserEmail(t *testing.T) {
	store := memory.NewMemoryStore()
	userId := "user123"
	refreshTokenHash := "hash123"
	mockEmail := "user@gmail.com"

	err := store.SaveRefreshTokenHash(userId, refreshTokenHash)
	require.NoError(t, err)

	t.Run("successful getting email", func(t *testing.T) {
		actualEmail, err := store.GetUserEmail(userId)
		require.NoError(t, err)
		require.Equal(t, mockEmail, actualEmail)
	})

	t.Run("invalid userId", func(t *testing.T) {
		invalidUserId := "user1234"
		_, err := store.GetUserEmail(invalidUserId)
		require.Error(t, err)
		require.ErrorContains(t, err, "refresh token hash is not found")
	})
}
