package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrSessionExpired = fmt.Errorf("session expired")
)

type Session struct {
	Password      string    `json:"password"`
	LastAccessed time.Time `json:"last_accessed"`
}

type sessionFile struct {
	Password      string `json:"password"`
	LastAccessed int64  `json:"last_accessed"`
}

func WriteSession(path string, session Session, ttl time.Duration) error {
	sf := sessionFile{
		Password:      session.Password,
		LastAccessed: time.Now().Unix(),
	}

	data, err := json.Marshal(sf)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}

	return nil
}

func ReadSession(path string) (Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("read session file: %w", err)
	}

	var sf sessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return Session{}, fmt.Errorf("unmarshal session: %w", err)
	}

	lastAccessed := time.Unix(sf.LastAccessed, 0)
	if time.Since(lastAccessed) > 30*time.Minute {
		return Session{}, ErrSessionExpired
	}

	return Session{
		Password:      sf.Password,
		LastAccessed: lastAccessed,
	}, nil
}

func RenewSession(path string, ttl time.Duration) error {
	session, err := ReadSession(path)
	if err != nil {
		return err
	}

	sf := sessionFile{
		Password:      session.Password,
		LastAccessed: time.Now().Unix(),
	}

	data, err := json.Marshal(sf)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}

	return nil
}

func GetSessionPath(dbPath string) string {
	return filepath.Join(os.TempDir(), "balde-session-"+hashPath(dbPath))
}

func hashPath(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:])[:16]
}