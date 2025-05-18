package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Set up test database
	var err error
	testDB, err = NewPostgresDB("postgres://postgres:postgres@localhost:5432/auth_test?sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Run migrations
	if err := RunMigrations("postgres://postgres:postgres@localhost:5432/auth_test?sslmode=disable", "file://../../migrations"); err != nil {
		panic(err)
	}

	code := m.Run()

	// Clean up
	testDB.Close()
	os.Exit(code)
}

func TestUserRepository_Integration(t *testing.T) {
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("create and get user", func(t *testing.T) {
		user := &User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)

		// Verify UUID format
		_, err = uuid.Parse(user.ID)
		require.NoError(t, err)

		// Get by username
		found, err := repo.GetByUsername(ctx, user.Username)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.PasswordHash, found.PasswordHash)

		// Get by email
		found, err = repo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)

		// Get by ID
		found, err = repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})
}

func TestTokenRepository_Integration(t *testing.T) {
	userRepo := NewUserRepository(testDB)
	tokenRepo := NewTokenRepository(testDB)
	ctx := context.Background()

	t.Run("create, get, and delete token", func(t *testing.T) {
		// Create a test user first
		user := &User{
			Username:     "tokenuser",
			Email:        "token@example.com",
			PasswordHash: "hashedpassword",
		}
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		token := &RefreshToken{
			UserID:    user.ID,
			Token:     "testtoken",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err = tokenRepo.Create(ctx, token)
		require.NoError(t, err)
		assert.NotEmpty(t, token.ID)

		// Verify UUID format
		_, err = uuid.Parse(token.ID)
		require.NoError(t, err)

		// Get token
		found, err := tokenRepo.Get(ctx, token.Token)
		require.NoError(t, err)
		assert.Equal(t, token.ID, found.ID)
		assert.Equal(t, token.UserID, found.UserID)
		assert.Equal(t, token.Token, found.Token)

		// Delete token
		err = tokenRepo.Delete(ctx, token.Token)
		require.NoError(t, err)

		// Try to get deleted token
		_, err = tokenRepo.Get(ctx, token.Token)
		assert.Error(t, err)
	})

	t.Run("delete all user tokens", func(t *testing.T) {
		// Create a test user first
		user := &User{
			Username:     "multitokenuser",
			Email:        "multitokens@example.com",
			PasswordHash: "hashedpassword",
		}
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		token1 := &RefreshToken{
			UserID:    user.ID,
			Token:     "token1",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		token2 := &RefreshToken{
			UserID:    user.ID,
			Token:     "token2",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err = tokenRepo.Create(ctx, token1)
		require.NoError(t, err)
		err = tokenRepo.Create(ctx, token2)
		require.NoError(t, err)

		// Delete all tokens for user
		err = tokenRepo.DeleteAllForUser(ctx, user.ID)
		require.NoError(t, err)

		// Try to get deleted tokens
		_, err = tokenRepo.Get(ctx, token1.Token)
		assert.Error(t, err)
		_, err = tokenRepo.Get(ctx, token2.Token)
		assert.Error(t, err)
	})
}
