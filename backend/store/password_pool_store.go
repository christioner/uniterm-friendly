package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/ys-ll/uniterm/backend/credentials"
)

const passwordPoolFileName = "passwords.json"

// PasswordAccount is a single saved credential in the password pool. Password
// holds the current secret — plaintext after Load, encrypted (enc:v1:) on disk
// via Save. The other fields are metadata and remain readable while locked.
type PasswordAccount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Notes     string `json:"notes,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// PasswordPoolData is the top-level structure of passwords.json.
type PasswordPoolData struct {
	Accounts []PasswordAccount `json:"accounts"`
}

// PasswordPoolStore persists the password pool to passwords.json, encrypting
// only the Password field of each account — the same model as IdentityStore. It
// is fail-closed: without a wired PasswordStore it refuses to save a plaintext
// secret, and decryption fails closed while the credential store is locked.
type PasswordPoolStore struct {
	configDir     string
	passwordStore PasswordStore // nil = refuse to write plaintext passwords
	mu            sync.Mutex    // serializes Save + Load password rewrites
}

func NewPasswordPoolStore(configDir string) (*PasswordPoolStore, error) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	return &PasswordPoolStore{configDir: configDir}, nil
}

// SetPasswordStore sets the external password store used to encrypt the
// Password field on save and decrypt it on load.
func (s *PasswordPoolStore) SetPasswordStore(ps PasswordStore) {
	s.passwordStore = ps
}

func (s *PasswordPoolStore) filePath() string {
	return filepath.Join(s.configDir, passwordPoolFileName)
}

// Save writes data to passwords.json, encrypting only the Password field of
// each account. It fails closed rather than persisting a plaintext password
// when no passwordStore is wired. Values already prefixed (credentials.
// IsEncrypted) are skipped so a save doesn't double-encrypt.
func (s *PasswordPoolStore) Save(data PasswordPoolData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accounts := make([]PasswordAccount, len(data.Accounts))
	copy(accounts, data.Accounts)

	for i := range accounts {
		if err := encryptSecretField(&accounts[i].Password, s.passwordStore); err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(PasswordPoolData{Accounts: accounts}, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(s.filePath(), jsonData, 0600)
}

// Load reads passwords.json, returning an empty pool when the file does not
// exist, quarantining+erroring on corrupt JSON, and decrypting the Password
// field of each account.
func (s *PasswordPoolStore) Load() (PasswordPoolData, error) {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return PasswordPoolData{Accounts: []PasswordAccount{}}, nil
		}
		return PasswordPoolData{}, err
	}
	var out PasswordPoolData
	if err := json.Unmarshal(data, &out); err != nil {
		quarantineCorrupt(s.filePath())
		return PasswordPoolData{}, err
	}
	if out.Accounts == nil {
		out.Accounts = []PasswordAccount{}
	}
	for i := range out.Accounts {
		acct := &out.Accounts[i]
		if acct.Password == "" || !credentials.IsEncrypted(acct.Password) || s.passwordStore == nil {
			continue
		}
		dec, err := s.passwordStore.Decrypt(acct.Password)
		if err != nil {
			return PasswordPoolData{}, err
		}
		acct.Password = dec
	}
	return out, nil
}