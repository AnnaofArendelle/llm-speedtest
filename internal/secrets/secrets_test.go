package secrets

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", ".secret")
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := c.Encrypt("sk-secret-value")
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Decrypt(blob)
	if err != nil || got != "sk-secret-value" {
		t.Fatalf("decrypt = %q %v", got, err)
	}
	// same key file -> same codec
	c2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := c2.Decrypt(blob); err != nil || got != "sk-secret-value" {
		t.Fatal("reload")
	}
	// different key fails
	other, _ := FromKey(make([]byte, 32))
	if _, err := other.Decrypt(blob); err == nil {
		t.Fatal("wrong key must fail")
	}
	if _, err := c.Decrypt([]byte("x")); err == nil {
		t.Fatal("short")
	}
	if err := os.WriteFile(path, []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("invalid key file must fail")
	}
	var nilc *Codec
	if _, err := nilc.Encrypt("x"); !errors.Is(err, ErrNoKey) {
		t.Fatal("nil codec")
	}
	if Mask("") != "" || Mask("short") != "••••" || Mask("sk-1234567890") != "sk-1••••7890" {
		t.Fatalf("mask %q", Mask("sk-1234567890"))
	}
}
