package services_test

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"auth_service/internal/services"
	"auth_service/internal/storage/storage_mocks"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestGenerateTokens тестирует функцию генерации токенов в сервисе аутентификации.
func TestGenerateTokens(t *testing.T) {
	mockStorage := storage_mocks.NewStorageInterface(t)
	authService := services.NewAuthService(mockStorage)

	t.Run("successful generating", func(t *testing.T) {
		userId := "user123"
		ip := "127.0.0.1"

		mockStorage.On("SaveRefreshTokenHash", userId, mock.Anything).Return(nil)

		tokensPair, err := authService.GenerateTokens(userId, ip)

		require.NoError(t, err)
		require.NotNil(t, tokensPair)
		require.NotEmpty(t, tokensPair.AccessToken)
		require.NotEmpty(t, tokensPair.RefreshToken)
		mockStorage.AssertCalled(t, "SaveRefreshTokenHash", userId, mock.Anything)
	})

	t.Run("storage failure on SaveRefreshTokenHash", func(t *testing.T) {
		userId := "user123"
		ip := "127.0.0.1"

		mockStorage.ExpectedCalls = nil
		mockStorage.On("SaveRefreshTokenHash", userId, mock.Anything).Return(errors.New("storage error"))

		tokensPair, err := authService.GenerateTokens(userId, ip)

		require.Error(t, err)
		require.Nil(t, tokensPair)
		require.ErrorContains(t, err, "failed to save refresh token")
		mockStorage.AssertCalled(t, "SaveRefreshTokenHash", userId, mock.Anything)
	})
}

// TestRefreshTokens тестирует функцию обновления токенов в сервисе аутентификации.
func TestRefreshTokens(t *testing.T) {
	mockStorage := storage_mocks.NewStorageInterface(t)
	authService := services.NewAuthService(mockStorage)

	testTokenClaims := &services.TokenClaims{
		UserId:   "userId",
		IssuedIp: "127.0.0.1",
		Jti:      "jti123",
	}
	ip := "127.0.0.1"
	config.Secret = "secret"

	accessToken, err := testTokenClaims.GenAccessToken()
	require.NoError(t, err)
	refreshToken, err := testTokenClaims.GenRefreshToken()
	require.NoError(t, err)

	oldTokenPair := &entities.TokensPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

    refrTokenHash, err := services.GenBcryptHash(refreshToken)
    require.NoError(t, err)

	t.Run("successful refresh", func(t *testing.T) {
		mockStorage.On("GetRefreshTokenHash", mock.Anything).Return(refrTokenHash, nil)
		mockStorage.On("UpdateRefreshTokenHash", mock.Anything, mock.Anything).Return(nil)

		newTokensPair, err := authService.RefreshTokens(ip, oldTokenPair)

		require.NoError(t, err)
		require.NotNil(t, newTokensPair)
		require.NotEmpty(t, newTokensPair.AccessToken)
		require.NotEmpty(t, newTokensPair.RefreshToken)
		mockStorage.AssertCalled(t, "GetRefreshTokenHash", mock.Anything)
		mockStorage.AssertCalled(t, "UpdateRefreshTokenHash", mock.Anything, mock.Anything)
	})

	t.Run("storage failure on GetRefreshTokenHash", func(t *testing.T) {
		mockStorage.ExpectedCalls = nil
		mockStorage.On("GetRefreshTokenHash", mock.Anything).Return("", errors.New("storage error"))

		tokensPair, err := authService.RefreshTokens(ip, oldTokenPair)

		require.Error(t, err)
		require.Nil(t, tokensPair)
		require.ErrorContains(t, err, "failed to get refresh token")
		mockStorage.AssertCalled(t, "GetRefreshTokenHash", mock.Anything)
	})

	t.Run("storage failure on UpdateRefreshTokenHash", func(t *testing.T) {
        mockStorage.ExpectedCalls = nil
		mockStorage.On("GetRefreshTokenHash", mock.Anything).Return(refrTokenHash, nil)
		mockStorage.On("UpdateRefreshTokenHash", mock.Anything, mock.Anything).Return(errors.New("storage error"))

		tokensPair, err := authService.RefreshTokens(ip, oldTokenPair)

		require.Error(t, err)
		require.Nil(t, tokensPair)
		require.ErrorContains(t, err, "failed to update refresh token")
		mockStorage.AssertCalled(t, "GetRefreshTokenHash", mock.Anything)
		mockStorage.AssertCalled(t, "UpdateRefreshTokenHash", mock.Anything, mock.Anything)
	})
}
