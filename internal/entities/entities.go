package entities

import "time"

// TokensPair представляет пару токенов: access и refresh.
// Используется для передачи токенов между клиентом и сервером.
type TokensPair struct {
	AccessToken  string `json:"access_token"`  // Access-токен для авторизации пользователя в системе.
	RefreshToken string `json:"refresh_token"` // Refresh-токен для обновления Access-токена.
}

// AccessTokenClaims представляет claims для access токена (JWT).
// Используется для проверки подлинности и срока действия access токена.
type AccessTokenClaims struct {
	Jti       string    // Уникальный идентификатор токена (JWT ID).
	CreatedAt time.Time // Время создания токена.
	UserId    string    // Идентификатор пользователя, которому принадлежит токен.
	ExpiredAt time.Time // Время истечения срока действия токена.
}

// RefreshTokenRecord представляет запись о refresh токене в базе данных.
// Используется для хранения и проверки refresh токена.
type RefreshTokenRecord struct {
	Jti       string    `db:"jti"`        // Уникальный идентификатор токена (JWT ID).
	CreatedAt time.Time `db:"created_at"` // Время создания токена.
	ExpiredAt time.Time `db:"expired_at"` // Время истечения срока действия токена.
	IssuedIp  string    `db:"issued_ip"`  // IP-адрес, с которого был выдан токен.
	TokenHash string    `db:"token_hash"` // Хэш refresh-токена для безопасного хранения.
}
