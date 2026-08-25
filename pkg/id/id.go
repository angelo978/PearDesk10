// Package id handles complete Tor Orion host IDs.
package id

import (
    "errors"
    "os"
    "path/filepath"
    "strings"
)

// Percorso del file ID persistente
func idFilePath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "peardesk", "id.txt")
}

// Load returns the saved ID if present and valid
func Load() (string, error) {
    path := idFilePath()

    data, err := os.ReadFile(path)
    if err != nil {
        return "", errors.New("ID non trovato")
    }

    id := strings.TrimSpace(string(data))
    if !Valid(id) {
        return "", errors.New("ID salvato non valido")
    }

    return id, nil
}

// Save writes the ID to disk
func Save(id string) error {
    if !Valid(id) {
        return errors.New("ID non valido")
    }

    path := idFilePath()
    os.MkdirAll(filepath.Dir(path), 0755)

    return os.WriteFile(path, []byte(id), 0644)
}

// ParseHost accepts only a complete v3 Onion address. Short IDs, IP addresses,
// public-IP IDs, and transport-specific aliases are intentionally rejected.
func ParseHost(value string) (string, error) {
    value = strings.TrimSpace(value)
    if len(value) != 62 || !strings.HasSuffix(value, ".onion") {
        return "", errors.New("ID Onion completo non valido")
    }
    for _, r := range strings.TrimSuffix(value, ".onion") {
        if !((r >= 'a' && r <= 'z') || (r >= '2' && r <= '7')) {
            return "", errors.New("ID Onion completo non valido")
        }
    }
    return value, nil
}

func Valid(value string) bool { _, err := ParseHost(value); return err == nil }

// Generate is deliberately empty: only Tor Orion may generate the ID.
func Generate() string { return "" }
