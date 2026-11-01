// Package secrets encrypts API keys at rest with AES-256-GCM using a master
// key stored next to the database.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Codec encrypts and decrypts short secrets.
type Codec struct {
	aead cipher.AEAD
}

// ErrNoKey is returned when the store has no master key.
var ErrNoKey = errors.New("no master key")

// Load reads the master key from path, creating a random one when missing.
func Load(path string) (*Codec, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(key)), 0o600); err != nil {
			return nil, err
		}
		return FromKey(key)
	}
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(string(trimSpace(raw)))
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("master key file %s is invalid", path)
	}
	return FromKey(key)
}

func trimSpace(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	return b
}

// FromKey builds a codec from a 32-byte key.
func FromKey(key []byte) (*Codec, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Codec{aead: aead}, nil
}

// Encrypt returns nonce||ciphertext.
func (c *Codec) Encrypt(plain string) ([]byte, error) {
	if c == nil {
		return nil, ErrNoKey
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return append(nonce, c.aead.Seal(nil, nonce, []byte(plain), nil)...), nil
}

// Decrypt reverses Encrypt.
func (c *Codec) Decrypt(blob []byte) (string, error) {
	if c == nil {
		return "", ErrNoKey
	}
	ns := c.aead.NonceSize()
	if len(blob) < ns {
		return "", errors.New("ciphertext too short")
	}
	plain, err := c.aead.Open(nil, blob[:ns], blob[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Mask renders a key for display: first 4 and last 4 characters.
func Mask(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "••••"
	}
	return key[:4] + "••••" + key[len(key)-4:]
}
