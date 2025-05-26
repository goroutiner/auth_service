package database

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Database представляет собой структуру для работы с базой данных
// и выполнения операций с таблицей refresh_tokens.
type Database struct {
	db *sqlx.DB // db — соединение с базой данных через sqlx
}

// NewDatabaseStore возвращает объект базы данных для работы с таблицами.
func NewDatabaseStore(db *sqlx.DB) *Database {
	return &Database{db: db}
}

// NewDatabaseConection устанавливает соединение с базой данных и создает таблицу refresh_tokens, если она не существует.
// Возвращает объект подключения к базе данных или ошибку.
func NewDatabaseConection(psqlUrl string) (*sqlx.DB, error) {
	if psqlUrl == "" {
		return nil, fmt.Errorf("'PSQL_URL' environment variable is empty")
	}

	db, err := sqlx.Open("pgx", psqlUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
	jti TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	expired_at  TIMESTAMP NOT NULL,
	issued_ip TEXT NOT NULL,
	token_hash TEXT UNIQUE NOT NULL
    );
	
	CREATE INDEX IF NOT EXISTS user_id__hash_indx ON refresh_tokens USING hash (user_id);
	CREATE INDEX IF NOT EXISTS token__hash_indx ON refresh_tokens USING hash (token_hash);
	`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table in database: %w", err)
	}

	log.Println("Starting the database ...")

	return db, err
}

// SaveRefreshTokenRecord сохраняет хэш refresh-токена для указанного пользователя.
// Если у пользователя уже максимальное количество токенов, удаляет самый старый токен.
// Возвращает ошибку, если токен с таким jti уже существует.
func (d *Database) SaveRefreshTokenRecord(userId string, refreshTokenRecord *entities.RefreshTokenRecord) error {
	maxTokensPerUser, err := strconv.Atoi(config.MaxTokensPerUser)
	if err != nil {
		return fmt.Errorf("env 'MAX_TOKENS_PER_USER' is not number: %w", err)
	}
	if err := d.checkActiveTokens(userId, maxTokensPerUser); err != nil {
		if !strings.Contains(err.Error(), "exceeding the limit for userID") {
			return fmt.Errorf("failed to check active tokens userID: '%s': %w", userId, err)
		}
		log.Println(err)
		if err := d.deleteOldestRefreshToken(userId); err != nil {
			return fmt.Errorf("failed to delete oldest refresh token for userID: '%s': %w", userId, err)
		}
		log.Printf("the latest token has been deleted due to exceeding the limit for userID: '%s'\n", userId)
	}

	query := `
	INSERT INTO refresh_tokens 
	(jti, user_id, created_at, expired_at, issued_ip, token_hash) 
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := d.db.Exec(query, refreshTokenRecord.Jti, userId, refreshTokenRecord.CreatedAt,
		refreshTokenRecord.ExpiredAt, refreshTokenRecord.IssuedIp, refreshTokenRecord.TokenHash); err != nil {
		return fmt.Errorf("failed to insert row into 'refresh_tokens' for userID: '%s': %w", userId, err)
	}

	return nil
}

// UpdateRefreshTokenRecord обновляет refresh-токен пользователя по старому jti.
// Удаляет старый токен и добавляет новый.
// Если запись не найдена, возвращает ошибку.
func (d *Database) UpdateRefreshTokenRecord(oldJti, userId string, newRefreshTokenRecord *entities.RefreshTokenRecord) error {
	query := `
	UPDATE refresh_tokens
	SET jti = $1, created_at = $2, expired_at = $3, issued_ip = $4, token_hash = $5  
    WHERE jti = $6 AND user_id = $7
	`

	result, err := d.db.Exec(query, newRefreshTokenRecord.Jti, newRefreshTokenRecord.CreatedAt, newRefreshTokenRecord.ExpiredAt,
		newRefreshTokenRecord.IssuedIp, newRefreshTokenRecord.TokenHash, oldJti, userId)
	if err != nil {
		return fmt.Errorf("failed to update row from 'refresh_tokens' for for userID: '%s': %w", userId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for userID: '%s': %w", userId, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated for userID: '%s': jti '%s' not found", userId, oldJti)
	}

	return nil
}

// GetRefreshTokenRecord возвращает record токена по jti и userId.
// Если запись не найдена, возвращает ошибку.
func (d *Database) GetRefreshTokenRecord(jti, userId string) (*entities.RefreshTokenRecord, error) {
	refreshTokenRecord := &entities.RefreshTokenRecord{}
	query := `
	SELECT jti, created_at, expired_at, issued_ip, token_hash
	FROM refresh_tokens 
    WHERE jti = $1 AND user_id = $2
	`

	if err := d.db.Get(refreshTokenRecord, query, jti, userId); err != nil {
		return nil, fmt.Errorf("failed to select claims from 'refresh_tokens' for jti: '%s': %w", jti, err)
	}

	return refreshTokenRecord, nil
}

// GetUserEmail возвращает email пользователя (в данном случае моковые данные).
// Если email не найден, возвращает ошибку.
func (d *Database) GetUserEmail(userId string) (string, error) {
	mockEmail := "user@gmail.com"

	return mockEmail, nil
}

// checkActiveTokens проверяет количество активных refresh токенов для пользователя.
// Если лимит превышен, возвращает специальную ошибку.
func (d *Database) checkActiveTokens(userId string, maxTokensPerUser int) error {
	var countOfSessions int
	query := `
	SELECT COUNT(*) 
	FROM refresh_tokens
	WHERE user_id = $1
	`

	if err := d.db.Get(&countOfSessions, query, userId); err != nil {
		return fmt.Errorf("failed to select count of active sessions from 'refresh_tokens' for userID: %s : %w", userId, err)
	}
	if countOfSessions == maxTokensPerUser {
		return fmt.Errorf("exceeding the limit for userID: '%s'", userId)
	}

	return nil
}

// deleteOldestRefreshToken удаляет самый старый refresh-токен пользователя.
// Если количество токенов не равно лимиту, возвращает ошибку.
func (d *Database) deleteOldestRefreshToken(userId string) error {
	query := `
	DELETE FROM refresh_tokens
	WHERE jti = (
		SELECT jti FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	)
	`

	result, err := d.db.Exec(query, userId)
	if err != nil {
		return fmt.Errorf("failed to delete row from 'refresh_tokens' for userID: '%s': %w", userId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for userID: '%s': %w", userId, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}
