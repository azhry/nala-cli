package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	APIBaseURL string `json:"apiBaseURL,omitempty"`
	SVCBaseURL string `json:"svcBaseURL,omitempty"`
}

type SettingsStore struct {
	Path string
}

func NewSettingsStore() (*SettingsStore, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}
	return &SettingsStore{Path: filepath.Join(root, "config.json")}, nil
}

func (s *SettingsStore) Load() (Settings, error) {
	if s == nil || strings.TrimSpace(s.Path) == "" {
		return Settings{}, errors.New("settings store is not configured")
	}
	payload, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return Settings{}, nil
	}
	if err != nil {
		return Settings{}, err
	}
	var settings Settings
	if err := json.Unmarshal(payload, &settings); err != nil {
		return Settings{}, errors.New("config file is invalid")
	}
	return settings, nil
}

func (s *SettingsStore) Save(settings Settings) error {
	if s == nil || strings.TrimSpace(s.Path) == "" {
		return errors.New("settings store is not configured")
	}
	settings.APIBaseURL = strings.TrimSpace(settings.APIBaseURL)
	settings.SVCBaseURL = strings.TrimSpace(settings.SVCBaseURL)
	if settings.APIBaseURL != "" {
		if err := ValidateBaseURL(settings.APIBaseURL); err != nil {
			return fmt.Errorf("api URL: %w", err)
		}
	}
	if settings.SVCBaseURL != "" {
		if err := ValidateBaseURL(settings.SVCBaseURL); err != nil {
			return fmt.Errorf("svc URL: %w", err)
		}
	}
	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
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

func ValidateBaseURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must be an http(s) URL without query or fragment")
	}
	return nil
}
