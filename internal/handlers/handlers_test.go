package handlers

import (
	"auth_service/internal/entities"
	"auth_service/internal/services/service_mocks"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestGenerateTokens проверяет работу обработчика GenerateTokens.
func TestGenerateTokens(t *testing.T) {
	mockService := service_mocks.NewAuthServiceInterface(t)
	handler := RegisterAuthHandler(mockService)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/{user_id}", handler.GenerateTokens())
	baseURL := "/api/auth"
	tokensPair := entities.TokensPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	t.Run("successful generated tokens", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		userId := "123"
		testURL := fmt.Sprintf("%s/%s", baseURL, userId)

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		respRec := httptest.NewRecorder()

		mockService.On("GenerateTokens", mock.Anything, mock.Anything).Return(&tokensPair, nil)
		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusOK, respRec.Code)

		var actualTokensPair entities.TokensPair

		err := json.NewDecoder(respRec.Body).Decode(&actualTokensPair)
		require.NoErrorf(t, err, "Ошибка парсинга JSON-ответа: %v", err)
		require.Equal(t, tokensPair, actualTokensPair)

		mockService.AssertCalled(t, "GenerateTokens", userId, req.RemoteAddr)
	})
	t.Run("user_id in URL is empty", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		testURL := baseURL

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusNotFound, respRec.Code)

		mockService.AssertNotCalled(t, "GenerateTokens")
	})
	t.Run("user_id is invalid", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		userId := "123qwerty"
		testURL := fmt.Sprintf("%s/%s", baseURL, userId)

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusBadRequest, respRec.Code)
		require.Contains(t, respRec.Body.String(), "user_id in URL must be integer")

		mockService.AssertNotCalled(t, "GenerateTokens")
	})
	t.Run("IP address is empty", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		userId := "123"
		testURL := fmt.Sprintf("%s/%s", baseURL, userId)

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		req.RemoteAddr = ""
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusBadRequest, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Client IP address is missing")

		mockService.AssertNotCalled(t, "GenerateTokens")
	})
	t.Run("unsuccessful generated tokens", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		userId := "123"
		testURL := fmt.Sprintf("%s/%s", baseURL, userId)

		req := httptest.NewRequest(http.MethodGet, testURL, nil)
		respRec := httptest.NewRecorder()

		mockService.On("GenerateTokens", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("some error"))
		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusUnauthorized, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Failed to generate token pair")

		mockService.AssertCalled(t, "GenerateTokens", userId, req.RemoteAddr)
	})
}

// // TestRefreshTokens проверяет работу обработчика RefreshTokens.
func TestRefreshTokens(t *testing.T) {
	mockService := service_mocks.NewAuthServiceInterface(t)
	handler := RegisterAuthHandler(mockService)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/refresh", handler.RefreshTokens())
	baseURL := "/api/auth/refresh"
	tokensPair := entities.TokensPair{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
	}
	refreshedTokens := entities.TokensPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}
	t.Run("successful refreshed tokens", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		reqBody, err := json.Marshal(tokensPair)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader(reqBody))
		respRec := httptest.NewRecorder()

		mockService.On("RefreshTokens", mock.Anything, mock.Anything).Return(&refreshedTokens, nil)
		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusOK, respRec.Code)

		var actualTokensPair entities.TokensPair

		err = json.NewDecoder(respRec.Body).Decode(&actualTokensPair)
		require.NoErrorf(t, err, "Ошибка парсинга JSON-ответа: %v", err)
		require.Equal(t, refreshedTokens, actualTokensPair)

		mockService.AssertCalled(t, "RefreshTokens", req.RemoteAddr, &tokensPair)
	})

	t.Run("invalid request body", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		reqBody, err := json.Marshal(`invalid json`)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader(reqBody))
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusBadRequest, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Invalid JSON")

		mockService.AssertNotCalled(t, "RefreshTokens")
	})

	t.Run("empty refresh token", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		emptyTokensPair := entities.TokensPair{
			AccessToken:  "access-token",
			RefreshToken: "",
		}
		reqBody, err := json.Marshal(emptyTokensPair)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader(reqBody))
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusBadRequest, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Refresh token is required")

		mockService.AssertNotCalled(t, "RefreshTokens")
	})

	t.Run("empty access token", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		emptyTokensPair := entities.TokensPair{
			AccessToken:  "",
			RefreshToken: "refresh-token",
		}
		reqBody, err := json.Marshal(emptyTokensPair)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader(reqBody))
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusBadRequest, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Access token is required")

		mockService.AssertNotCalled(t, "RefreshTokens")
	})
	t.Run("IP address is empty", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		reqBody, err := json.Marshal(tokensPair)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader(reqBody))
		req.RemoteAddr = ""
		respRec := httptest.NewRecorder()

		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusBadRequest, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Client IP address is missing")

		mockService.AssertNotCalled(t, "RefreshTokens")
	})
	t.Run("unsuccessful refreshed tokens", func(t *testing.T) {
		t.Cleanup(func() { mockService.ExpectedCalls = nil })

		reqBody, err := json.Marshal(tokensPair)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader(reqBody))
		respRec := httptest.NewRecorder()

		mockService.On("RefreshTokens", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("some error"))
		mux.ServeHTTP(respRec, req)
		require.Equal(t, http.StatusUnauthorized, respRec.Code)
		require.Contains(t, respRec.Body.String(), "Failed to refresh Token Pairs")

		mockService.AssertCalled(t, "RefreshTokens", req.RemoteAddr, &tokensPair)
	})
}
