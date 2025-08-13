package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/fortega2/real-time-chat/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

const (
	msgCreateUserFailed      = "CreateUser failed: %v"
	msgGetUserByIDFailed     = "GetUserByID failed: %v"
	msgGetUserByUsernameFail = "GetUserByUsername failed: %v"
)

func initializeTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_username_password ON users (username, password);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
	return db
}

func TestCreateUser(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T) (*repository.Queries, func())
		params  repository.CreateUserParams
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Success",
			setup: func(t *testing.T) (*repository.Queries, func()) {
				db := initializeTestDB(t)
				return repository.New(db), func() { _ = db.Close() }
			},
			params: repository.CreateUserParams{Username: "alice", Password: "secret"},
		},
		{
			name: "DuplicateUsername",
			setup: func(t *testing.T) (*repository.Queries, func()) {
				db := initializeTestDB(t)
				q := repository.New(db)
				if _, err := q.CreateUser(context.Background(), repository.CreateUserParams{Username: "bob", Password: "p1"}); err != nil {
					t.Fatalf("seed CreateUser failed: %v", err)
				}
				return q, func() { _ = db.Close() }
			},
			params:  repository.CreateUserParams{Username: "bob", Password: "p2"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { runCreateUserCase(t, tc) })
	}
}

func runCreateUserCase(t *testing.T, tc struct {
	name    string
	setup   func(t *testing.T) (*repository.Queries, func())
	params  repository.CreateUserParams
	wantErr bool
}) {
	t.Helper()
	q, cleanup := tc.setup(t)
	defer cleanup()

	u, err := q.CreateUser(context.Background(), tc.params)
	if tc.wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf(msgCreateUserFailed, err)
	}
	assertUserCreated(t, u, tc.params.Username, tc.params.Password)
}

func TestGetUserByID(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(t *testing.T) (*repository.Queries, int64, func())
		wantFound bool
	}

	tests := []testCase{
		{
			name: "Found",
			setup: func(t *testing.T) (*repository.Queries, int64, func()) {
				db := initializeTestDB(t)
				q := repository.New(db)
				u, err := q.CreateUser(context.Background(), repository.CreateUserParams{Username: "carol", Password: "pw"})
				if err != nil {
					t.Fatalf(msgCreateUserFailed, err)
				}
				return q, u.ID, func() { _ = db.Close() }
			},
			wantFound: true,
		},
		{
			name: "NotFound",
			setup: func(t *testing.T) (*repository.Queries, int64, func()) {
				db := initializeTestDB(t)
				return repository.New(db), 9999, func() { _ = db.Close() }
			},
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { runGetUserByIDCase(t, tc) })
	}
}

func runGetUserByIDCase(t *testing.T, tc struct {
	name      string
	setup     func(t *testing.T) (*repository.Queries, int64, func())
	wantFound bool
}) {
	t.Helper()
	q, id, cleanup := tc.setup(t)
	defer cleanup()

	got, err := q.GetUserByID(context.Background(), id)
	if !tc.wantFound {
		if err == nil {
			t.Fatal("expected error for non-existent ID, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf(msgGetUserByIDFailed, err)
	}
	if got.ID != id {
		t.Fatalf("expected ID %d, got %d", id, got.ID)
	}
}

func TestGetUserByUsername(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(t *testing.T) (*repository.Queries, string, func())
		wantFound bool
	}

	tests := []testCase{
		{
			name: "Found",
			setup: func(t *testing.T) (*repository.Queries, string, func()) {
				db := initializeTestDB(t)
				q := repository.New(db)
				_, err := q.CreateUser(context.Background(), repository.CreateUserParams{Username: "dave", Password: "pw2"})
				if err != nil {
					t.Fatalf(msgCreateUserFailed, err)
				}
				return q, "dave", func() { _ = db.Close() }
			},
			wantFound: true,
		},
		{
			name: "NotFound",
			setup: func(t *testing.T) (*repository.Queries, string, func()) {
				db := initializeTestDB(t)
				return repository.New(db), "nope", func() { _ = db.Close() }
			},
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { runGetUserByUsernameCase(t, tc) })
	}
}

func runGetUserByUsernameCase(t *testing.T, tc struct {
	name      string
	setup     func(t *testing.T) (*repository.Queries, string, func())
	wantFound bool
}) {
	t.Helper()
	q, username, cleanup := tc.setup(t)
	defer cleanup()

	got, err := q.GetUserByUsername(context.Background(), username)
	if !tc.wantFound {
		if err == nil {
			t.Fatal("expected error for non-existent username, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf(msgGetUserByUsernameFail, err)
	}
	if got.Username != username {
		t.Fatalf("expected username %q, got %q", username, got.Username)
	}
}

func assertUserCreated(t *testing.T, u repository.User, wantUsername, wantPassword string) {
	t.Helper()
	if u.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if u.Username != wantUsername {
		t.Fatalf("expected username %q, got %q", wantUsername, u.Username)
	}
	if u.Password != wantPassword {
		t.Fatalf("expected password to round-trip, got %q", u.Password)
	}
	if u.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}
