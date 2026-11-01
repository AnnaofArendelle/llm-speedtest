package web

import (
	"io/fs"
	"testing"
)

func TestDist(t *testing.T) {
	dist := Dist()
	if _, err := fs.Stat(dist, "."); err != nil {
		t.Fatalf("dist root: %v", err)
	}
	_, err := fs.Stat(dist, "index.html")
	if built := err == nil; built != Built() {
		t.Fatalf("Built() = %v but index.html stat err = %v", Built(), err)
	}
	if Built() {
		raw, err := fs.ReadFile(dist, "index.html")
		if err != nil || len(raw) == 0 {
			t.Fatalf("read index.html: %v", err)
		}
	}
}
