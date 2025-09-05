package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/fortega2/real-time-chat/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

const (
	msgCreateChannelFailed  = "CreateChannel failed: %v"
	msgGetChannelByIDFailed = "GetChannelByID failed: %v"
	msgGetAllChannelsFailed = "GetAllChannels failed: %v"
	msgDeleteChannelFailed  = "DeleteChannel failed: %v"
)

func initializeTestDBWithChannels(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	usersSchema := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL UNIQUE,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_users_username_password ON users (username, password);
    `
	if _, err := db.Exec(usersSchema); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Create channels table
	channelsSchema := `
    CREATE TABLE IF NOT EXISTS channels (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        created_by INTEGER NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
    );
    `
	if _, err := db.Exec(channelsSchema); err != nil {
		t.Fatalf("failed to create channels table: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, q *repository.Queries) repository.User {
	t.Helper()
	user, err := q.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "testuser",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestCreateChannel(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T) (*repository.Queries, int64, func())
		params  repository.CreateChannelParams
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Success",
			setup: func(t *testing.T) (*repository.Queries, int64, func()) {
				db := initializeTestDBWithChannels(t)
				q := repository.New(db)
				user := createTestUser(t, q)
				return q, user.ID, func() { _ = db.Close() }
			},
			params: repository.CreateChannelParams{
				Name:        "general",
				Description: sql.NullString{String: "General discussion", Valid: true},
			},
		},
		{
			name: "Success - No Description",
			setup: func(t *testing.T) (*repository.Queries, int64, func()) {
				db := initializeTestDBWithChannels(t)
				q := repository.New(db)
				user := createTestUser(t, q)
				return q, user.ID, func() { _ = db.Close() }
			},
			params: repository.CreateChannelParams{
				Name:        "random",
				Description: sql.NullString{Valid: false},
			},
		},
		{
			name: "Duplicate Channel Name",
			setup: func(t *testing.T) (*repository.Queries, int64, func()) {
				db := initializeTestDBWithChannels(t)
				q := repository.New(db)
				user := createTestUser(t, q)
				if _, err := q.CreateChannel(context.Background(), repository.CreateChannelParams{
					Name:        "duplicate",
					Description: sql.NullString{String: "First channel", Valid: true},
					CreatedBy:   user.ID,
				}); err != nil {
					t.Fatalf("seed CreateChannel failed: %v", err)
				}
				return q, user.ID, func() { _ = db.Close() }
			},
			params: repository.CreateChannelParams{
				Name:        "duplicate",
				Description: sql.NullString{String: "Second channel", Valid: true},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runCreateChannelCase(t, tc)
		})
	}
}

func runCreateChannelCase(t *testing.T, tc struct {
	name    string
	setup   func(t *testing.T) (*repository.Queries, int64, func())
	params  repository.CreateChannelParams
	wantErr bool
}) {
	t.Helper()
	q, userID, cleanup := tc.setup(t)
	defer cleanup()

	tc.params.CreatedBy = userID
	ch, err := q.CreateChannel(context.Background(), tc.params)
	if tc.wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf(msgCreateChannelFailed, err)
	}
	assertChannelCreated(t, ch, tc.params.Name, tc.params.Description, userID)
}

func TestGetAllChannels(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (*repository.Queries, int, func())
		wantCount int
	}{
		{
			name:      "Empty Result",
			setup:     setupEmptyChannels,
			wantCount: 0,
		},
		{
			name:      "Multiple Channels",
			setup:     setupMultipleChannels,
			wantCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runGetAllChannelsTest(t, tc.setup, tc.wantCount)
		})
	}
}

func setupEmptyChannels(t *testing.T) (*repository.Queries, int, func()) {
	db := initializeTestDBWithChannels(t)
	return repository.New(db), 0, func() { _ = db.Close() }
}

func setupMultipleChannels(t *testing.T) (*repository.Queries, int, func()) {
	db := initializeTestDBWithChannels(t)
	q := repository.New(db)
	user := createTestUser(t, q)

	channels := []string{"general", "random", "dev"}
	createTestChannels(t, q, user.ID, channels)

	return q, len(channels), func() { _ = db.Close() }
}

func createTestChannels(t *testing.T, q *repository.Queries, userID int64, channelNames []string) {
	for _, name := range channelNames {
		if _, err := q.CreateChannel(context.Background(), repository.CreateChannelParams{
			Name:        name,
			Description: sql.NullString{String: name + " channel", Valid: true},
			CreatedBy:   userID,
		}); err != nil {
			t.Fatalf("failed to create channel %s: %v", name, err)
		}
	}
}

func runGetAllChannelsTest(t *testing.T, setup func(t *testing.T) (*repository.Queries, int, func()), expectedCount int) {
	q, _, cleanup := setup(t)
	defer cleanup()

	channels, err := q.GetAllChannels(context.Background())
	if err != nil {
		t.Fatalf(msgGetAllChannelsFailed, err)
	}

	if len(channels) != expectedCount {
		t.Fatalf("expected %d channels, got %d", expectedCount, len(channels))
	}

	validateChannelOrdering(t, channels)
}

func validateChannelOrdering(t *testing.T, channels []repository.GetAllChannelsRow) {
	if len(channels) <= 1 {
		return
	}

	for i := 0; i < len(channels)-1; i++ {
		if channels[i].CreatedAt.Before(channels[i+1].CreatedAt) {
			t.Error("channels should be ordered by created_at DESC")
			break
		}
	}
}

func TestDeleteChannel(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*repository.Queries, repository.DeleteChannelParams, func())
		wantErr bool
	}{
		{
			name:    "Success",
			setup:   setupSuccessfulDelete,
			wantErr: false,
		},
		{
			name:    "Wrong Creator",
			setup:   setupWrongCreatorDelete,
			wantErr: false,
		},
		{
			name:    "Non-existent Channel",
			setup:   setupNonExistentChannelDelete,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runDeleteChannelTest(t, tc.setup, tc.wantErr)
		})
	}
}

func setupSuccessfulDelete(t *testing.T) (*repository.Queries, repository.DeleteChannelParams, func()) {
	db := initializeTestDBWithChannels(t)
	q := repository.New(db)
	user := createTestUser(t, q)

	chId, err := q.CreateChannel(context.Background(), repository.CreateChannelParams{
		Name:        "todelete",
		Description: sql.NullString{String: "Channel to delete", Valid: true},
		CreatedBy:   user.ID,
	})
	if err != nil {
		t.Fatalf(msgCreateChannelFailed, err)
	}

	params := repository.DeleteChannelParams{
		ID:        chId,
		CreatedBy: user.ID,
	}

	return q, params, func() { _ = db.Close() }
}

func setupWrongCreatorDelete(t *testing.T) (*repository.Queries, repository.DeleteChannelParams, func()) {
	db := initializeTestDBWithChannels(t)
	q := repository.New(db)
	user1 := createTestUser(t, q)

	user2, err := q.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "user2",
		Password: "pass2",
	})
	if err != nil {
		t.Fatalf("failed to create second user: %v", err)
	}

	chId := createChannelForUser(t, q, user1.ID, "protected", "Protected channel")

	params := repository.DeleteChannelParams{
		ID:        chId,
		CreatedBy: user2.ID,
	}

	return q, params, func() { _ = db.Close() }
}

func setupNonExistentChannelDelete(t *testing.T) (*repository.Queries, repository.DeleteChannelParams, func()) {
	db := initializeTestDBWithChannels(t)
	q := repository.New(db)
	user := createTestUser(t, q)

	params := repository.DeleteChannelParams{
		ID:        9999,
		CreatedBy: user.ID,
	}

	return q, params, func() { _ = db.Close() }
}

func createChannelForUser(t *testing.T, q *repository.Queries, userID int64, name, description string) int64 {
	chId, err := q.CreateChannel(context.Background(), repository.CreateChannelParams{
		Name:        name,
		Description: sql.NullString{String: description, Valid: true},
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf(msgCreateChannelFailed, err)
	}
	return chId
}

func runDeleteChannelTest(t *testing.T, setup func(t *testing.T) (*repository.Queries, repository.DeleteChannelParams, func()), wantErr bool) {
	q, params, cleanup := setup(t)
	defer cleanup()

	err := q.DeleteChannel(context.Background(), params)

	if wantErr && err == nil {
		t.Fatal("expected error, got nil")
	}

	if !wantErr && err != nil {
		t.Fatalf(msgDeleteChannelFailed, err)
	}
}

func assertChannelCreated(t *testing.T, chId int64, wantName string, wantDescription sql.NullString, wantCreatedBy int64) {
	t.Helper()
	if chId == 0 {
		t.Fatal("expected non-zero ID")
	}
}
