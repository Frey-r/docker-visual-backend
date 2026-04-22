package auth

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// UserStore handles user persistence.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new user store with SQLite.
func NewUserStore(dbPath string) (*UserStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create users table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return &UserStore{db: db}, nil
}

// Close closes the database connection.
func (s *UserStore) Close() error {
	return s.db.Close()
}

// CreateUser creates a new user.
func (s *UserStore) CreateUser(username, passwordHash string) (*User, error) {
	// Check if user exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Insert user
	result, err := s.db.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        id,
		Username:  username,
		CreatedAt: time.Now(),
	}, nil
}

// GetUserByUsername retrieves a user by username.
func (s *UserStore) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	var createdAt string
	err := s.db.QueryRow(
		"SELECT id, username, password, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &createdAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return user, nil
}

// GetUserByID retrieves a user by ID.
func (s *UserStore) GetUserByID(id int64) (*User, error) {
	user := &User{}
	var createdAt string
	err := s.db.QueryRow(
		"SELECT id, username, password, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &createdAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return user, nil
}

// UserCount returns the number of users.
func (s *UserStore) UserCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
