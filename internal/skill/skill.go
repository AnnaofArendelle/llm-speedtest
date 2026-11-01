// Package skill installs and exports the embedded AI skill package.
package skill

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnnaofArendelle/llm-speedtest/skills"
)

// Name is the skill directory name.
const Name = "llm-speedtest"

// Files returns the skill file system rooted at the skill directory.
func Files() fs.FS {
	sub, err := fs.Sub(skills.FS, Name)
	if err != nil {
		panic(err)
	}
	return sub
}

// Manifest returns SKILL.md.
func Manifest() string {
	raw, err := fs.ReadFile(Files(), "SKILL.md")
	if err != nil {
		return ""
	}
	return string(raw)
}

// Validate checks the SKILL.md frontmatter and the presence of key files.
func Validate(fsys fs.FS) error {
	raw, err := fs.ReadFile(fsys, "SKILL.md")
	if err != nil {
		return fmt.Errorf("SKILL.md missing: %w", err)
	}
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return errors.New("SKILL.md must start with YAML frontmatter")
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return errors.New("SKILL.md frontmatter is not terminated")
	}
	fm := text[4 : 4+end]
	var name, desc string
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
		if strings.HasPrefix(line, "description:") {
			desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	if name != Name {
		return fmt.Errorf("frontmatter name %q must be %q", name, Name)
	}
	if len(desc) < 50 || len(desc) > 1024 {
		return fmt.Errorf("description must be 50..1024 characters, got %d", len(desc))
	}
	for _, f := range []string{"scripts/install.sh", "references/cli.md", "references/metrics.md", "assets/report-template.md"} {
		if _, err := fs.Stat(fsys, f); err != nil {
			return fmt.Errorf("missing %s", f)
		}
	}
	return nil
}

// TargetDir resolves an install target to a directory.
func TargetDir(target string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch strings.ToLower(target) {
	case "", "claude", "claude-code":
		return filepath.Join(home, ".claude", "skills", Name), nil
	case "codex":
		return filepath.Join(home, ".codex", "skills", Name), nil
	case "cursor":
		return filepath.Join(home, ".cursor", "skills", Name), nil
	case "project":
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, ".claude", "skills", Name), nil
	}
	if filepath.Base(target) == Name {
		return filepath.Clean(target), nil
	}
	return filepath.Join(filepath.Clean(target), Name), nil
}

// Install copies the skill into the target directory.
func Install(target string, force bool) (string, error) {
	dir, err := TargetDir(target)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err == nil && !force {
		return dir, fmt.Errorf("%s already contains a skill; use --force to overwrite", dir)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	err = fs.WalkDir(Files(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(dst, 0o750)
		}
		raw, err := fs.ReadFile(Files(), p)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(p, ".sh") {
			mode = 0o755
		}
		return os.WriteFile(dst, raw, mode)
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// ExportZip writes the skill package as a zip archive containing the
// llm-speedtest/ directory.
func ExportZip(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return WriteZip(f)
}

// WriteZip streams the skill package as a zip archive.
func WriteZip(w io.Writer) error {
	zw := zip.NewWriter(w)
	err := fs.WalkDir(Files(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := fs.ReadFile(Files(), p)
		if err != nil {
			return err
		}
		hdr := &zip.FileHeader{Name: Name + "/" + p, Method: zip.Deflate}
		if strings.HasSuffix(p, ".sh") {
			hdr.SetMode(0o755)
		} else {
			hdr.SetMode(0o644)
		}
		fw, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		_, err = fw.Write(raw)
		return err
	})
	if err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}
