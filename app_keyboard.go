package main

import (
	"errors"

	"github.com/ys-ll/uniterm/backend/store"
)

// Keyboard shortcuts live in their own synced file (keyboard.json) so they can
// follow the user across machines, unlike the rest of settings.json which stays
// device-local (theme, paths, UI state).

// LoadKeyboardSettings returns the custom bindings. On first run (no keyboard.json
// yet) it migrates the user's existing shortcuts out of the settings blob and
// persists them, so nothing is lost when they start editing.
func (a *App) LoadKeyboardSettings() (map[string]store.KeyBinding, error) {
	if a.keyboardStore == nil {
		return nil, errors.New("keyboard store not initialized")
	}
	kb, err := a.keyboardStore.Load()
	if err != nil {
		return nil, err
	}
	if kb != nil {
		return kb, nil
	}

	// First run: migrate from the settings blob, then persist.
	kb = map[string]store.KeyBinding{}
	if a.settingsStore != nil {
		if s, err := a.settingsStore.Load(); err == nil && s.Keyboard != nil {
			kb = s.Keyboard
		}
	}
	_ = a.keyboardStore.Save(kb)
	return kb, nil
}

// SaveKeyboardSettings persists the custom bindings to keyboard.json.
func (a *App) SaveKeyboardSettings(kb map[string]store.KeyBinding) error {
	if a.keyboardStore == nil {
		return errors.New("keyboard store not initialized")
	}
	return a.keyboardStore.Save(kb)
}