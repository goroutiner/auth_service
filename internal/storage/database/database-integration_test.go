package database_test

import (
	"auth_service/internal/storage/database"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	psqlC  testcontainers.Container
	testDb *sqlx.DB
	store  *database.Database
)

// TestMain производит соединения с БД и создает таблицу для хранения тестовых данных.
func TestMain(m *testing.M) {
	var err error

	ctx := context.Background()
	buildContext, err := filepath.Abs("./docker")
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v\n", err)
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: buildContext,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	psqlC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	host, err := psqlC.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := psqlC.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatal(err)
	}

	psqlUrl := fmt.Sprintf("postgres://user:password@%s:%s/test_db?sslmode=disable", host, mappedPort.Port())

	testDb, err = database.NewDatabaseConection(psqlUrl)
	if err != nil {
		log.Fatalf("failed connection to the database: %v\n", err)
	}

	store = database.NewDatabaseStore(testDb)

	code := m.Run()

	testcontainers.TerminateContainer(psqlC)
	testDb.Close()
	os.Exit(code)
}

// TestSaveRefreshTokenHash проверяет сохранение хэша refresh токена в БД.
func TestSaveRefreshTokenHash(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

	userId := "user123"
	refrTokenHash := "hash123"

	err := store.SaveRefreshTokenHash(userId, refrTokenHash)
	require.NoError(t, err)

	var count int
	err = testDb.Get(&count, "SELECT COUNT(*) FROM refresh_tokens WHERE user_id=$1 AND token_hash=$2", userId, refrTokenHash)
	require.NoError(t, err)
	require.Equal(t, 1, count, "refresh token's hash was not saved in the database")
}

// TestGetRefreshTokenHash проверяет получение хэша refresh токена из БД.
func TestGetRefreshTokenHash(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

	userId := "user123"
	mockEmail := "user@gmail.com"
	refrTokenHash := "hash123"

	_, err := testDb.Exec("INSERT INTO refresh_tokens (user_id, email, token_hash) VALUES ($1, $2, $3)", userId, mockEmail, refrTokenHash)
	require.NoError(t, err)

	t.Run("successful getting token", func(t *testing.T) {
		userId := "user123"
		result, err := store.GetRefreshTokenHash(userId)
		require.NoError(t, err)
		require.Equal(t, refrTokenHash, result)
	})

	t.Run("invalid userId", func(t *testing.T) {
		invalidUserId := "user1234"
		result, err := store.GetRefreshTokenHash(invalidUserId)
		require.ErrorContains(t, err, "no record found for user_id")
		require.Empty(t, result)
	})
}

// TestUpdateRefreshTokenHash проверяет обновление хэша refresh токена в БД.
func TestUpdateRefreshTokenHash(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

	userId := "user123"
	mockEmail := "user@gmail.com"
	oldRefrTokenHash := "oldhash456"
	newRefrTokenHash := "newhash456"

	_, err := testDb.Exec("INSERT INTO refresh_tokens (user_id, email, token_hash) VALUES ($1, $2, $3)", userId, mockEmail, oldRefrTokenHash)
	require.NoError(t, err)

	t.Run("successful updating token", func(t *testing.T) {
		userId := "user123"
		err = store.UpdateRefreshTokenHash(userId, newRefrTokenHash)
		require.NoError(t, err)

		var updatedHash string
		err = testDb.Get(&updatedHash, "SELECT token_hash FROM refresh_tokens WHERE user_id=$1", userId)
		require.NoError(t, err)
		require.Equal(t, newRefrTokenHash, updatedHash, "Хэш токена обновления не обновился в БД")
	})

	t.Run("invalid userId", func(t *testing.T) {
		var updatedHash string
		invalidUserId := "user1234"
		err = store.UpdateRefreshTokenHash(invalidUserId, newRefrTokenHash)
		require.ErrorContains(t, err, "no rows updated")
		require.Empty(t, updatedHash)
	})
}

// TestGetUserEmail проверяет получение email пользователя из БД.
func TestGetUserEmail(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

	userId := "user123"
	mockEmail := "user@gmail.com"
	refrTokenHash := "hash123"

	_, err := testDb.Exec("INSERT INTO refresh_tokens (user_id, email, token_hash) VALUES ($1, $2, $3)", userId, mockEmail, refrTokenHash)
	require.NoError(t, err)

	t.Run("successful getting email", func(t *testing.T) {
		result, err := store.GetUserEmail(userId)
		require.NoError(t, err)
		require.Equal(t, mockEmail, result)
	})

	t.Run("invalid userId", func(t *testing.T) {
		invalidUserId := "user1234"
		result, err := store.GetUserEmail(invalidUserId)
		assert.ErrorContains(t, err, "no record found for user_id")
		require.Empty(t, result)
	})
}

// truncateTable удаляет все записи из указанной таблицы в БД.
func truncateTable(spaceName string, t *testing.T) {
	query := "TRUNCATE TABLE refresh_tokens"
	_, err := testDb.Exec(query)
	require.NoError(t, err, "Failed to truncate table: %s", spaceName)
}
