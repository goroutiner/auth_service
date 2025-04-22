package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Database представляет собой структуру для работы с базой данных
// и выполнения операций с таблицей refresh_tokens.
type Database struct {
	db *sqlx.DB
}

// NewDatabaseStore возвращает объект базы данных для работы с таблицами.
func NewDatabaseStore(db *sqlx.DB) *Database {
	return &Database{db: db}
}

// NewDatabaseConection устанавливает соединение с базой данных и создает таблицу refresh_tokens, если она не существует.
func NewDatabaseConection(psqlUrl string) (*sqlx.DB, error) {
	if psqlUrl == "" {
		return nil, fmt.Errorf("'PSQL_URL' environment variable is empty")
	}

	db, err := sqlx.Open("pgx", psqlUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	query := `CREATE TABLE IF NOT EXISTS refresh_tokens (
	user_id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    token_hash TEXT UNIQUE NOT NULL
    )`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table in database: %w", err)
	}

	log.Println("Starting the database ...")

	return db, err
}

// SaveRefreshTokenHash сохраняет хэш токена обновления для указанного пользователя.
func (d *Database) SaveRefreshTokenHash(userId, refrTokenHash string) error {
	mockEmail := "user@gmail.com"

	query := `INSERT INTO refresh_tokens 
	(user_id, email, token_hash) VALUES ($1, $2, $3)`

	if _, err := d.db.Exec(query, userId, mockEmail, refrTokenHash); err != nil {
		return fmt.Errorf("failed to insert 'user_id', 'email', 'token_hash' into 'refresh_tokens': %w", err)
	}

	return nil
}

// GetRefreshTokenHash возвращает хэш токена обновления для указанного пользователя.
func (d *Database) GetRefreshTokenHash(userId string) (string, error) {
	var refrTokenHash string
	query := `SELECT token_hash 
	FROM refresh_tokens 
	WHERE user_id = $1`

	if err := d.db.Get(&refrTokenHash, query, userId); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no record found for user_id '%s'", userId)
		}
		return "", fmt.Errorf("failed to select 'token_hash' from 'refresh_tokens': %w", err)
	}

	return refrTokenHash, nil
}

// UpdateRefreshTokenHash обновляет хэш токена обновления для указанного пользователя.
func (d *Database) UpdateRefreshTokenHash(userId, newRefrTokenHash string) error {
	query := `UPDATE refresh_tokens
    SET token_hash = $1
    WHERE user_id = $2`

	result, err := d.db.Exec(query, newRefrTokenHash, userId)
	if err != nil {
		return fmt.Errorf("failed to update 'token_hash' from 'refresh_tokens': %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated: user_id '%s' not found", userId)
	}

	return nil
}

// GetUserEmail возвращает email пользователя по его идентификатору.
func (d *Database) GetUserEmail(userId string) (string, error) {
	var userEmail string
	query := `SELECT email 
	FROM refresh_tokens 
	WHERE user_id = $1`

	if err := d.db.Get(&userEmail, query, userId); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no record found for user_id '%s'", userId)
		}
		return "", fmt.Errorf("failed to select 'email' from 'refresh_tokens': %w", err)
	}

	return userEmail, nil
}
