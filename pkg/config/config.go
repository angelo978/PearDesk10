// Package config loads and saves PearDesk user configuration.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// HistoryEntry records a previous remote-desktop connection.
type HistoryEntry struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	LastConnected    time.Time `json:"last_connected"`
	RememberPassword bool      `json:"remember_password"`
	Password         string    `json:"password,omitempty"`
}

// Config is the persisted PearDesk configuration.
type Config struct {
	HostID       string         `json:"host_id"`                // complete Tor Orion hostname
	HostPassword string         `json:"host_password"`          // empty = no auth
	TorDataDir   string         `json:"tor_data_dir,omitempty"` // persistent Onion key directory
	History      []HistoryEntry `json:"history"`
	Language     string         `json:"language,omitempty"`
}

func dataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".peardesk"
	}
	return filepath.Join(home, ".peardesk")
}

func configPath() string { return filepath.Join(dataDir(), "config.json") }

// Load reads config from disk; returns a default config on first run.
func Load() (*Config, error) {
	_ = os.MkdirAll(dataDir(), 0700)
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, nil
	}
	return &cfg, nil
}

// Save writes config to disk.
func (c *Config) Save() error {
	_ = os.MkdirAll(dataDir(), 0700)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

// AddOrUpdateHistory upserts a history entry (most-recent first, max 50).
func (c *Config) AddOrUpdateHistory(id, name, password string, remember bool) {
	for i, h := range c.History {
		if h.ID == id {
			c.History[i].Name = name
			c.History[i].LastConnected = time.Now()
			c.History[i].RememberPassword = remember
			if remember {
				c.History[i].Password = password
			} else {
				c.History[i].Password = ""
			}
			return
		}
	}
	entry := HistoryEntry{
		ID:               id,
		Name:             name,
		LastConnected:    time.Now(),
		RememberPassword: remember,
	}
	if remember {
		entry.Password = password
	}
	c.History = append([]HistoryEntry{entry}, c.History...)
	if len(c.History) > 50 {
		c.History = c.History[:50]
	}
}

// GetHistoryPassword returns the saved password for id (if any).
func (c *Config) GetHistoryPassword(id string) (string, bool) {
	for _, h := range c.History {
		if h.ID == id && h.RememberPassword {
			return h.Password, true
		}
	}
	return "", false
}

// RemoveHistory deletes the history entry with the given id.
func (c *Config) RemoveHistory(id string) {
	filtered := c.History[:0]
	for _, h := range c.History {
		if h.ID != id {
			filtered = append(filtered, h)
		}
	}
	c.History = filtered
}
