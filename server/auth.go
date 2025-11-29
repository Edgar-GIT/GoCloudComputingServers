package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Token represents an authentication token
type Token struct {
	Value     string
	Username  string
	ExpiresAt time.Time
}

// User represents a user
type User struct {
	Username string
	Password string
}

// AuthManager manages authentication
type AuthManager struct {
	tokens    map[string]*Token
	users     map[string]*User
	mu        sync.RWMutex
	credsFile string // Path to USER_CREDS.json file
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(credsFilePath string) *AuthManager {
	am := &AuthManager{
		tokens:    make(map[string]*Token),
		users:     make(map[string]*User),
		credsFile: credsFilePath,
	}

	// Create default admin user
	am.users["admin"] = &User{
		Username: "admin",
		Password: "admin",
	}

	// Load credentials from file if it exists
	am.LoadUsersFromFile()

	// Save credentials (to create file if it doesn't exist)
	am.SaveUsersToFile()

	return am
}

// GenerateToken generates a new token for the user
func (am *AuthManager) GenerateToken(username string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	am.mu.Lock()
	defer am.mu.Unlock()

	// Store token with 24 hour expiration
	am.tokens[token] = &Token{
		Value:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

// ValidateToken validates a token
func (am *AuthManager) ValidateToken(token string) (*Token, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	t, exists := am.tokens[token]
	if !exists {
		return nil, errors.New("invalid token")
	}

	// Check if token has expired
	if time.Now().After(t.ExpiresAt) {
		delete(am.tokens, token)
		return nil, errors.New("token expired")
	}

	return t, nil
}

// RevokeToken removes a token
func (am *AuthManager) RevokeToken(token string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.tokens, token)
}

// CleanupExpiredTokens removes expired tokens
func (am *AuthManager) CleanupExpiredTokens() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	for token, t := range am.tokens {
		if now.After(t.ExpiresAt) {
			delete(am.tokens, token)
		}
	}
}

// Authenticate verifies user credentials
func (am *AuthManager) Authenticate(username, password string) bool {
	if username == "" || password == "" {
		return false
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.users[username]
	if !exists {
		return false
	}

	return user.Password == password
}

// CreateUser creates a new user
func (am *AuthManager) CreateUser(username, password string) error {
	if username == "" || password == "" {
		return errors.New("username and password are required")
	}

	// Validate username (cannot be admin)
	if username == "admin" {
		return errors.New("cannot create a user with the name 'admin'")
	}

	// Validate username characters (only letters, numbers and underscore)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return errors.New("username contains invalid characters")
		}
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if user already exists
	if _, exists := am.users[username]; exists {
		return errors.New("user already exists")
	}

	// Create new user
	am.users[username] = &User{
		Username: username,
		Password: password,
	}

	// Save credentials to file
	if err := am.SaveUsersToFile(); err != nil {
		// Log error but don't fail user creation
		// In production, should log this properly
		return err
	}

	return nil
}

// LoadUsersFromFile loads users from JSON file
func (am *AuthManager) LoadUsersFromFile() error {
	if am.credsFile == "" {
		return nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(am.credsFile); os.IsNotExist(err) {
		return nil // File doesn't exist, use only admin
	}

	// Read file
	data, err := os.ReadFile(am.credsFile)
	if err != nil {
		return err
	}

	// Parse JSON
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	// Load users (except admin which already exists)
	for _, user := range users {
		if user.Username != "admin" {
			userCopy := user // Create copy to avoid pointer issue
			am.users[user.Username] = &userCopy
		}
	}

	return nil
}

// SaveUsersToFile saves users to JSON file
func (am *AuthManager) SaveUsersToFile() error {
	if am.credsFile == "" {
		return nil
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	// Convert map to slice
	users := make([]User, 0, len(am.users))
	for _, user := range am.users {
		users = append(users, *user)
	}

	// Convert to JSON
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(am.credsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file
	return os.WriteFile(am.credsFile, data, 0600) // Permissions: only owner can read/write
}

// UserExists checks if a user exists
func (am *AuthManager) UserExists(username string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	_, exists := am.users[username]
	return exists
}
