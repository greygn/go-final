package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Set up test database
	var err error
	testDB, err = NewPostgresDB("postgres://postgres:postgres@localhost:5432/forum_test?sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Run migrations
	if err := RunMigrations("postgres://postgres:postgres@localhost:5432/forum_test?sslmode=disable", "file://../../migrations"); err != nil {
		panic(err)
	}

	code := m.Run()

	// Clean up
	testDB.Close()
	os.Exit(code)
}

func TestMessageRepository_Integration(t *testing.T) {
	repo := NewMessageRepository(testDB)
	ctx := context.Background()

	t.Run("create and list messages", func(t *testing.T) {
		message := &Message{
			UserID:    1,
			Username:  "testuser",
			Content:   "Hello, World!",
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, message)
		require.NoError(t, err)
		assert.NotZero(t, message.ID)

		// List messages
		messages, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, message.ID, messages[0].ID)
		assert.Equal(t, message.UserID, messages[0].UserID)
		assert.Equal(t, message.Username, messages[0].Username)
		assert.Equal(t, message.Content, messages[0].Content)
	})

	t.Run("delete old messages", func(t *testing.T) {
		oldMessage := &Message{
			UserID:    1,
			Username:  "testuser",
			Content:   "Old message",
			CreatedAt: time.Now().Add(-time.Hour),
		}
		newMessage := &Message{
			UserID:    1,
			Username:  "testuser",
			Content:   "New message",
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, oldMessage)
		require.NoError(t, err)
		err = repo.Create(ctx, newMessage)
		require.NoError(t, err)

		// Delete messages older than 30 minutes
		err = repo.DeleteOld(ctx, time.Minute*30)
		require.NoError(t, err)

		// List remaining messages
		messages, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, newMessage.ID, messages[0].ID)
	})
}
