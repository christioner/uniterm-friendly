package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ys-ll/uniterm/backend/store"
)

// ---- Password pool (账号密码池) ----
//
// A user-configurable pool of credentials that can be injected into a running
// terminal when it asks for a password (sudo / passwd / ssh). Entries reuse
// the unified credential master key (enc:v1: AES-GCM via the credential
// store), so unlocking the pool is the same as unlocking connection passwords.
// Plaintext passwords are decrypted only on demand (GetPasswordAccountPassword)
// and never returned in the list binding.

// ListPasswordAccounts returns the pool metadata only — the Password field is
// cleared so plaintext never reaches the frontend list.
func (a *App) ListPasswordAccounts() (store.PasswordPoolData, error) {
	if a.passwordPoolStore == nil {
		return store.PasswordPoolData{Accounts: []store.PasswordAccount{}}, nil
	}
	data, err := a.passwordPoolStore.Load()
	if err != nil {
		return store.PasswordPoolData{}, err
	}
	for i := range data.Accounts {
		data.Accounts[i].Password = ""
	}
	return data, nil
}

// SavePasswordAccount creates or updates an account (upsert by ID). New
// accounts get a generated ID and created timestamp.
func (a *App) SavePasswordAccount(acct store.PasswordAccount) error {
	if a.passwordPoolStore == nil {
		return errors.New("password pool store not initialized")
	}
	data, err := a.passwordPoolStore.Load()
	if err != nil {
		return err
	}
	if acct.ID == "" {
		acct.ID = uuid.NewString()
		acct.CreatedAt = time.Now().Unix()
	}
	found := false
	for i := range data.Accounts {
		if data.Accounts[i].ID == acct.ID {
			// Editing with an empty password means "keep the current one" —
			// data has been decrypted by Load(), so accounts[i].Password is
			// the plaintext to preserve.
			if acct.Password == "" {
				acct.Password = data.Accounts[i].Password
			}
			data.Accounts[i] = acct
			found = true
			break
		}
	}
	if !found {
		data.Accounts = append(data.Accounts, acct)
	}
	return a.passwordPoolStore.Save(data)
}

// DeletePasswordAccount removes an account by ID.
func (a *App) DeletePasswordAccount(id string) error {
	if a.passwordPoolStore == nil {
		return errors.New("password pool store not initialized")
	}
	data, err := a.passwordPoolStore.Load()
	if err != nil {
		return err
	}
	out := make([]store.PasswordAccount, 0, len(data.Accounts))
	for _, acct := range data.Accounts {
		if acct.ID != id {
			out = append(out, acct)
		}
	}
	data.Accounts = out
	return a.passwordPoolStore.Save(data)
}

// GetPasswordAccountPassword returns the decrypted plaintext password for an
// account. It is gated on the credential store being unlocked (Decrypt fails
// closed otherwise). The result is meant to feed a single inject-in-terminal
// call and must not be retained.
func (a *App) GetPasswordAccountPassword(id string) (string, error) {
	if a.passwordPoolStore == nil {
		return "", errors.New("password pool store not initialized")
	}
	data, err := a.passwordPoolStore.Load()
	if err != nil {
		return "", err
	}
	for _, acct := range data.Accounts {
		if acct.ID == id {
			return acct.Password, nil
		}
	}
	return "", errors.New("password account not found")
}

// InjectPassword writes password into the given session's stdin — the same
// path as SessionWrite — so a selected pool password never has to live in
// frontend state.
func (a *App) InjectPassword(sessionID, password string) error {
	if a.sessionManager == nil {
		return errors.New("session manager not initialized")
	}
	s, ok := a.sessionManager.Get(sessionID)
	if !ok {
		return errors.New("session not found: " + sessionID)
	}
	return s.Write([]byte(password))
}