package database_test

import (
	"auth_service/internal/config"
	"auth_service/internal/entities"
	"auth_service/internal/storage/database"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
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
	config.MaxTokensPerUser = "5"
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

// TestSaveRefreshTokenRecord проверяет сохранение refresh токена в БД, включая ошибочные кейсы.
func TestSaveRefreshTokenRecord(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

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

	actualRefreshTokenRecord := &entities.RefreshTokenRecord{}
	query := `SELECT jti, created_at, expired_at, issued_ip, token_hash FROM refresh_tokens WHERE user_id=$1 AND jti=$2`
	err = testDb.Get(actualRefreshTokenRecord, query, userId, refreshTokenRecord.Jti)
	require.NoError(t, err)
	require.Equal(t, refreshTokenRecord, actualRefreshTokenRecord)

	t.Run("duplicate add", func(t *testing.T) {
		err := store.SaveRefreshTokenRecord(userId, refreshTokenRecord)
		require.Error(t, err)
		require.ErrorContains(t, err, "ERROR: duplicate key")
	})
}

// TestUpdateRefreshTokenRecord проверяет обновление refresh токена в БД, включая ошибочные кейсы.
func TestUpdateRefreshTokenRecord(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

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

	actualRefreshTokenRecord := &entities.RefreshTokenRecord{}
	query := `SELECT jti, created_at, expired_at, issued_ip, token_hash FROM refresh_tokens WHERE user_id=$1 AND jti=$2`
	err = testDb.Get(actualRefreshTokenRecord, query, userId, newRefreshTokenRecord.Jti)
	require.NoError(t, err)
	require.Equal(t, newRefreshTokenRecord, actualRefreshTokenRecord)

	t.Run("update non-existent", func(t *testing.T) {
		err := store.UpdateRefreshTokenRecord("not_exist_jti", userId, newRefreshTokenRecord)
		require.Error(t, err)
		require.ErrorContains(t, err, "not found")
	})
}

// TestGetRefreshTokenRecord проверяет получение refresh токена из БД, включая ошибочные кейсы.
func TestGetRefreshTokenRecord(t *testing.T) {
	t.Cleanup(func() { truncateTable("refresh_tokens", t) })

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

	actualRefreshTokenRecord, err := store.GetRefreshTokenRecord(refreshTokenRecord.Jti, userId)
	require.NoError(t, err)
	require.Equal(t, refreshTokenRecord, actualRefreshTokenRecord)

	t.Run("get non-existent", func(t *testing.T) {
		_, err := store.GetRefreshTokenRecord("not_exist_jti", userId)
		require.Error(t, err)
		require.ErrorContains(t, err, "sql: no rows in result set")
	})
}

// truncateTable удаляет все записи из указанной таблицы в БД.
func truncateTable(spaceName string, t *testing.T) {
	query := "TRUNCATE TABLE refresh_tokens"
	_, err := testDb.Exec(query)
	require.NoError(t, err, "Failed to truncate table: %s", spaceName)
}
