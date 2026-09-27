package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const keyboardStoreFileName = "keyboard.json"

// KeyboardStore persists the user's custom keyboard shortcuts to keyboard.json —
// the only device-independent slice of settings that participates in cloud sync
// (settings.json itself stays local: theme, paths, UI state). KeyBinding is
// declared in settings_store.go. No secret fields, so no PasswordStore wiring.
type KeyboardStore struct {
	configDir string
	mu        sync.Mutex
}

func NewKeyboardStore(configDir string) (*KeyboardStore, error) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	return &KeyboardStore{configDir: configDir}, nil
}

func (s *KeyboardStore) filePath() string {
	return filepath.Join(s.configDir, keyboardStoreFileName)
}

// Load returns the saved bindings, or (nil, nil) when the file is absent so the
// caller can fall back to another source (migration).
func (s *KeyboardStore) Load() (map[string]KeyBinding, error) {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var kb map[string]KeyBinding
	if err := json.Unmarshal(data, &kb); err != nil {
		return nil, err
	}
	return kb, nil
}

func (s *KeyboardStore) Save(kb map[string]KeyBinding) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if kb == nil {
		kb = map[string]KeyBinding{}
	}
	// Marshal a sorted struct? map order is irrelevant to JSON consumers; keep
	// output deterministic enough for sync comparison by sorting later if needed.
	data, err := json.MarshalIndent(kb, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(s.filePath(), data, 0600)
}