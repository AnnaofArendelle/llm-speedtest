package skill

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateInstallExport(t *testing.T) {
	if err := Validate(Files()); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(Manifest(), "---\nname: llm-speedtest") {
		t.Fatal("manifest")
	}
	dir := t.TempDir()
	got, err := Install(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(dir, Name) {
		t.Fatalf("dir = %s", got)
	}
	if _, err := os.Stat(filepath.Join(got, "scripts", "install.sh")); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" { // NTFS has no executable bit
		st, _ := os.Stat(filepath.Join(got, "scripts", "install.sh"))
		if st.Mode()&0o111 == 0 {
			t.Fatal("scripts must be executable")
		}
	}
	if _, err := Install(dir, false); err == nil {
		t.Fatal("second install without --force must fail")
	}
	if _, err := Install(dir, true); err != nil {
		t.Fatal(err)
	}
	if d, _ := TargetDir("claude"); !strings.HasSuffix(d, filepath.Join(".claude", "skills", Name)) {
		t.Fatalf("claude target = %s", d)
	}
	if d, _ := TargetDir(filepath.Join(dir, Name)); d != filepath.Join(dir, Name) {
		t.Fatalf("explicit target = %s", d)
	}
	zipPath := filepath.Join(dir, "skill.zip")
	if err := ExportZip(zipPath); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["llm-speedtest/SKILL.md"] || !names["llm-speedtest/references/metrics.md"] {
		t.Fatalf("zip entries = %v", names)
	}
}
