package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	GameDir  string `json:"game_directory"`
	PlutoDir string `json:"plutonium_directory"`
	Map      string `json:"map"`
	TestMenu bool   `json:"test_menu"`
}

func saveConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".survival-*")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	// Go's Windows rename replaces an existing ordinary file.
	return os.Rename(temp, path)
}

func gameValid(path string) bool {
	info, err := os.Stat(filepath.Join(path, "zone", "english"))
	return err == nil && info.IsDir() && fileExists(filepath.Join(path, "main", "iw_00.iwd"))
}

func fileExists(path string) bool { s, e := os.Stat(path); return e == nil && s.Mode().IsRegular() }

func launchArgs(cfg Config) ([]string, error) {
	if !regexp.MustCompile(`^mp_[a-z0-9_]+$`).MatchString(cfg.Map) {
		return nil, fmt.Errorf("Ongeldige mapnaam: %q", cfg.Map)
	}
	test := "0"
	if cfg.TestMenu {
		test = "1"
	}
	return []string{"iw5mp", cfg.GameDir, "-lan", "+set", "fs_game", "mods/survival", "+set", "survival_test_menu", test, "+load_dsr", "survival", "+map", cfg.Map}, nil
}

// Reject traversal and symlinked destination paths before touching any files.
func safeTarget(root, name string) (string, error) {
	if strings.Contains(name, "\\") || strings.Contains(name, ":") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("Onveilig archiefpad: %s", name)
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("Onveilig archiefpad: %s", name)
	}
	target := filepath.Join(root, clean)
	for cur := target; ; cur = filepath.Dir(cur) {
		info, err := os.Lstat(cur)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("Symlink niet toegestaan: %s", cur)
		}
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if cur == filepath.Dir(cur) {
			break
		}
	}
	return target, nil
}

type pendingFile struct {
	target  string
	data    []byte
	old     []byte
	existed bool
}

// Preflight all paths and preserve original bytes before modifying the installation.
// Backups are content-addressed, so reruns cannot overwrite an older backup.
func install(payload []byte, root, backup string) (int, error) {
	z, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return 0, err
	}
	pending := []pendingFile{}
	// Root zone fastfiles can load before fs_game mounts mods/survival. Make the
	// same original resource archive visible during that initial search phase.
	files := append([]*zip.File(nil), z.File...)
	var resources *zip.File
	for _, f := range z.File {
		if f.Name == "iw5/mods/survival/mod.iwd" {
			resources = f
		}
	}
	if resources != nil {
		for _, f := range z.File {
			if f.Name == "iw5/LB_Survival_Resources.iwd" {
				return 0, fmt.Errorf("Duplicate early resource destination")
			}
		}
		alias := *resources
		alias.Name = "iw5/LB_Survival_Resources.iwd"
		files = append(files, &alias)
	}
	for _, f := range files {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name, "iw5/") {
			return 0, fmt.Errorf("Onverwacht archiefpad: %s", f.Name)
		}
		target, err := safeTarget(root, strings.TrimPrefix(f.Name, "iw5/"))
		if err != nil {
			return 0, err
		}
		r, err := f.Open()
		if err != nil {
			return 0, err
		}
		data, err := io.ReadAll(io.LimitReader(r, 128*1024*1024+1))
		r.Close()
		if err != nil {
			return 0, err
		}
		if len(data) > 128*1024*1024 {
			return 0, fmt.Errorf("Bestand te groot: %s", f.Name)
		}
		old, err := os.ReadFile(target)
		exists := err == nil
		if err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		if exists && bytes.Equal(old, data) {
			continue
		}
		if exists {
			hash := sha256.Sum256(old)
			dest, err := safeTarget(backup, strings.TrimPrefix(f.Name, "iw5/")+fmt.Sprintf(".%x.bak", hash))
			if err != nil {
				return 0, err
			}
			if !fileExists(dest) {
				if err = atomicWrite(dest, old); err != nil {
					return 0, fmt.Errorf("Backup mislukt: %w", err)
				}
			}
		}
		pending = append(pending, pendingFile{target, data, old, exists})
	}
	for i, f := range pending {
		if err := atomicWrite(f.target, f.data); err != nil {
			failures := []string{}
			for j := i - 1; j >= 0; j-- {
				p := pending[j]
				var e error
				if p.existed {
					e = atomicWrite(p.target, p.old)
				} else {
					e = os.Remove(p.target)
				}
				if e != nil {
					failures = append(failures, e.Error())
				}
			}
			return 0, fmt.Errorf("Installatie mislukt: %w; herstelfouten: %v; backups: %s", err, failures, backup)
		}
	}
	return len(pending), nil
}
