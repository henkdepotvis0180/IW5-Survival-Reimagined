package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for name, data := range files {
		f, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = f.Write([]byte(data)); e != nil {
			t.Fatal(e)
		}
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}

func TestInstallBackupAndRepeat(t *testing.T) {
	root := t.TempDir()
	backup := t.TempDir()
	target := filepath.Join(root, "scripts", "test.gsc")
	if e := atomicWrite(target, []byte("original")); e != nil {
		t.Fatal(e)
	}
	payload := archive(t, map[string]string{"iw5/scripts/test.gsc": "updated", "iw5/mods/survival/test.ff": "asset"})
	count, e := install(payload, root, backup)
	if e != nil || count != 2 {
		t.Fatalf("%d %v", count, e)
	}
	b, e := os.ReadFile(target)
	if e != nil || string(b) != "updated" {
		t.Fatalf("%s %v", b, e)
	}
	found := 0
	filepath.WalkDir(backup, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			b, e := os.ReadFile(p)
			if e != nil || string(b) != "original" {
				t.Fatalf("bad backup %s %v", b, e)
			}
			found++
		}
		return nil
	})
	if found != 1 {
		t.Fatalf("backups: %d", found)
	}
	count, e = install(payload, root, backup)
	if e != nil || count != 0 {
		t.Fatalf("repeat: %d %v", count, e)
	}
}

func TestRejectTraversalBeforeChanges(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"iw5/../../escape", "iw5/C:/escape", "iw5/..\\escape", "foreign/test"} {
		_, e := install(archive(t, map[string]string{name: "bad"}), root, t.TempDir())
		if e == nil {
			t.Fatalf("accepted %s", name)
		}
	}
}

func TestRejectSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if e := os.Symlink(outside, filepath.Join(root, "scripts")); e != nil {
		t.Skip(e)
	}
	_, e := install(archive(t, map[string]string{"iw5/scripts/test.gsc": "bad"}), root, t.TempDir())
	if e == nil {
		t.Fatal("accepted symlink")
	}
	if fileExists(filepath.Join(outside, "test.gsc")) {
		t.Fatal("wrote outside installation")
	}
}

func TestLaunchArgs(t *testing.T) {
	args, e := launchArgs(Config{GameDir: `C:\Games\Call of Duty MW3`, Map: "mp_dome"})
	if e != nil {
		t.Fatal(e)
	}
	if args[1] != `C:\Games\Call of Duty MW3` {
		t.Fatal("split game path")
	}
	if !strings.Contains(strings.Join(args, " "), "survival_test_menu 0") {
		t.Fatal("test menu enabled by default")
	}
	if _, e = launchArgs(Config{Map: "mp_dome;quit"}); e == nil {
		t.Fatal("accepted command injection")
	}
	args, e = launchArgs(Config{GameDir: "game", Map: "mp_dome", TestMenu: true})
	if e != nil || !strings.Contains(strings.Join(args, " "), "survival_test_menu 1") {
		t.Fatal(args, e)
	}
}

func TestConfigAndGameValidation(t *testing.T) {
	root := t.TempDir()
	cfg := Config{GameDir: root, Map: "mp_dome"}
	path := filepath.Join(root, "state", "launcher.json")
	if e := saveConfig(path, cfg); e != nil {
		t.Fatal(e)
	}
	if gameValid(root) {
		t.Fatal("accepted incomplete game")
	}
	os.MkdirAll(filepath.Join(root, "zone", "english"), 0755)
	atomicWrite(filepath.Join(root, "main", "iw_00.iwd"), []byte("fixture"))
	if !gameValid(root) {
		t.Fatal("rejected valid directory layout")
	}
	// Config updates must replace the previous file too.
	cfg.TestMenu = true
	if e := saveConfig(path, cfg); e != nil {
		t.Fatal(e)
	}
}
