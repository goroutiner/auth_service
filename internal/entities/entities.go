package entities

// TokensPair представляет пару токенов: access и refresh.
// Используется для передачи токенов между клиентом и сервером.
type TokensPair struct {
	AccessToken  string `json:"access_token"`  // Access-токен для авторизации.
	RefreshToken string `json:"refresh_token"` // Refresh-токен для обновления Access-токена.
}

// Client представляет данные клиента.
type Client struct {
	Email            string // Email клиента.
	RefreshTokenHash string // Хэш refresh-токена клиента.
}
