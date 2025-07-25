package handlers

import (
	"auth_service/internal/entities"
	"auth_service/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// AuthHandler представляет обработчик для работы с аутентификацией.
type AuthHandler struct {
	service services.AuthServiceInterface
}

// RegisterAuthHandler регистрирует обработчик аутентификации.
func RegisterAuthHandler(service services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{service: service}
}

// GenerateTokens обрабатывает GET-запрос для генерации пары токенов (доступа и обновления).
// Ожидает user_id в параметрах пути и IP-адрес клиента.
// Возвращает JSON с новой парой токенов.
func (h *AuthHandler) GenerateTokens() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			newTokensPair *entities.TokensPair
			err           error
		)

		userId := r.PathValue("user_id")
		if _, err := strconv.Atoi(userId); err != nil {
			http.Error(w, "user_id in URL must be integer", http.StatusBadRequest)
			return
		}
		ip := r.RemoteAddr
		if ip == "" {
			log.Println("IP address is empty")
			http.Error(w, "Client IP address is missing", http.StatusBadRequest)
			return
		}

		newTokensPair, err = h.service.GenerateTokens(userId, ip)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to generate token pair", http.StatusUnauthorized)
			return
		}

		resp := entities.TokensPair{
			AccessToken:  newTokensPair.AccessToken,
			RefreshToken: newTokensPair.RefreshToken,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// RefreshTokens обрабатывает POST-запрос для обновления пары токенов.
// Ожидает JSON с access_token и refresh_token в теле запроса.
// Возвращает JSON с обновленной парой токенов.
func (s *AuthHandler) RefreshTokens() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req entities.TokensPair
			err error
		)

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println(err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		switch {
		case req.AccessToken == "":
			log.Println("access token is empty")
			http.Error(w, "Access token is required", http.StatusBadRequest)
			return
		case req.RefreshToken == "":
			log.Println("refresh token is empty")
			http.Error(w, "Refresh token is required", http.StatusBadRequest)
			return
		}

		ip := r.RemoteAddr
		if ip == "" {
			log.Println("IP address is empty")
			http.Error(w, "Client IP address is missing", http.StatusBadRequest)
			return
		}

		updTokensPair, err := s.service.RefreshTokens(ip, &req)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to refresh Token Pairs", http.StatusUnauthorized)
			return
		}

		resp := entities.TokensPair{
			AccessToken:  updTokensPair.AccessToken,
			RefreshToken: updTokensPair.RefreshToken,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
