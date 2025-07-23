package services

import (
	"auth_service/internal/config"
	"crypto/sha512"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func generateTestAccessToken(secret string, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(secret))
}

// TestParseAccessToken проверяет парсинг access токена JWT.
func TestParseAccessToken(t *testing.T) {
	t.Run("successful parsing", func(t *testing.T) {
		testJti := "test-jti"
		testUserId := "123"
		config.Secret = "test_secret"
		now := time.Now().Unix()
		exp := now + 3600

		claims := jwt.MapClaims{
			"jti":        testJti,
			"user_id":    testUserId,
			"created_at": now,
			"expired_at": exp,
		}
		accessToken, err := generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err, "failed to generate token: %v", err)

		accessTokenClaims, err := parseAccessToken(accessToken)
		require.NoError(t, err)

		require.Equal(t, testJti, accessTokenClaims.Jti)
		require.Equal(t, testUserId, accessTokenClaims.UserId)
		require.Equal(t, now, accessTokenClaims.CreatedAt.Unix())
		require.Equal(t, exp, accessTokenClaims.ExpiredAt.Unix())
	})
	t.Run("invalid token", func(t *testing.T) {
		config.Secret = "test_secret"
		invalidToken := "invalid.token.string"

		accessTokenClaims, err := parseAccessToken(invalidToken)
		require.ErrorContains(t, err, "failed to parse jwt")
		require.Nil(t, accessTokenClaims)
	})
	t.Run("missing claims", func(t *testing.T) {
		testJti := "test-jti"
		testUserId := "123"
		config.Secret = "test_secret"
		now := time.Now().Unix()
		claims := jwt.MapClaims{
			"jti": testJti,
			// missing user_id, created_at, expired_at
		}
		accessToken, err := generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err, "failed to generate token: %v", err)

		accessTokenClaims, err := parseAccessToken(accessToken)
		require.ErrorContains(t, err, "'user_id' in claims is empty")
		require.Nil(t, accessTokenClaims)

		claims = jwt.MapClaims{
			"jti":     testJti,
			"user_id": testUserId,
			// missing created_at, expired_at
		}

		accessToken, err = generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err, "failed to generate token: %v", err)

		accessTokenClaims, err = parseAccessToken(accessToken)
		require.ErrorContains(t, err, "'created_at' in claims is empty")
		require.Nil(t, accessTokenClaims)

		claims = jwt.MapClaims{
			"jti":        testJti,
			"user_id":    testUserId,
			"created_at": now,
			// missing expired_at
		}

		accessToken, err = generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err, "failed to generate token: %v", err)

		accessTokenClaims, err = parseAccessToken(accessToken)
		require.ErrorContains(t, err, "'expired_at' in claims is empty")
		require.Nil(t, accessTokenClaims)
	})
	t.Run("invalid claimTypes", func(t *testing.T) {
		validJti := "test-jti"
		validUserId := "123"
		config.Secret = "test_secret"
		now := time.Now().Unix()
		config.Secret = "test_secret"
		exp := now + 3600

		claims := jwt.MapClaims{
			"jti":        123, // should be string
			"user_id":    validUserId,
			"created_at": now,
			"expired_at": exp,
		}
		accessToken, err := generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err)

		accessTokenClaims, err := parseAccessToken(accessToken)
		require.ErrorContains(t, err, "failed to typecast 'jti' to string")
		require.Nil(t, accessTokenClaims)

		claims = jwt.MapClaims{
			"jti":        validJti,
			"user_id":    000, // should be string
			"created_at": now,
			"expired_at": exp,
		}
		accessToken, err = generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err)

		accessTokenClaims, err = parseAccessToken(accessToken)
		require.ErrorContains(t, err, "failed to typecast 'user_id' to string")
		require.Nil(t, accessTokenClaims)

		claims = jwt.MapClaims{
			"jti":        validJti,
			"user_id":    validUserId,
			"created_at": "not float64", // should be float64
			"expired_at": exp,
		}
		accessToken, err = generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err)

		accessTokenClaims, err = parseAccessToken(accessToken)
		require.ErrorContains(t, err, "failed to typecast 'created_at' to float64")
		require.Nil(t, accessTokenClaims)

		claims = jwt.MapClaims{
			"jti":        validJti,
			"user_id":    validUserId,
			"created_at": now,
			"expired_at": "not float64", // should be float64
		}
		accessToken, err = generateTestAccessToken(config.Secret, claims)
		require.NoError(t, err)

		accessTokenClaims, err = parseAccessToken(accessToken)
		require.ErrorContains(t, err, "failed to typecast 'expired_at' to float64")
		require.Nil(t, accessTokenClaims)
	})
}

// TestCheckRefreshToken проверяет валидацию refresh токена.
func TestCheckRefreshToken(t *testing.T) {
	t.Run("valid refresh token", func(t *testing.T) {
		refreshToken := "valid_refresh_token"
		hash := sha512.Sum512([]byte(refreshToken))
		validHash, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)
		require.NoError(t, err)

		err = checkRefreshToken(refreshToken, string(validHash))
		require.NoError(t, err)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshToken := "valid_refresh_token"
		invalidToken := "invalid_refresh_token"
		hash := sha512.Sum512([]byte(refreshToken))
		validHash, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)
		require.NoError(t, err)

		err = checkRefreshToken(invalidToken, string(validHash))
		require.ErrorContains(t, err, "refresh token hash is invalid")
	})

	t.Run("invalid hash format", func(t *testing.T) {
		refreshToken := "valid_refresh_token"
		invalidHash := "not_a_bcrypt_hash"

		err := checkRefreshToken(refreshToken, invalidHash)
		require.ErrorContains(t, err, "refresh token hash is invalid")
	})
}
