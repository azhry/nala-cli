package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type SessionStore struct {
	Path string
}

type sessionFile struct {
	Token string `json:"token"`
}

func NewSessionStore() (*SessionStore, error) {
	root := strings.TrimSpace(os.Getenv("NALA_CONFIG_DIR"))
	if root == "" {
		var err error
		root, err = os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(root, "nala")
	}
	return &SessionStore{Path: filepath.Join(root, "session.json")}, nil
}

func (s *SessionStore) Save(token string) error {
	if s == nil || strings.TrimSpace(s.Path) == "" {
		return errors.New("session store is not configured")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("session token is empty")
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	payload, err := json.Marshal(sessionFile{Token: token})
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	file, err := os.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func (s *SessionStore) Token() (string, error) {
	if s == nil || strings.TrimSpace(s.Path) == "" {
		return "", errors.New("session store is not configured")
	}
	payload, err := os.ReadFile(s.Path)
	if err != nil {
		return "", err
	}
	var stored sessionFile
	if err := json.Unmarshal(payload, &stored); err != nil {
		return "", errors.New("session file is invalid")
	}
	if strings.TrimSpace(stored.Token) == "" {
		return "", errors.New("session file has no token")
	}
	return stored.Token, nil
}
