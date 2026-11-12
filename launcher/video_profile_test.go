package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVideoProfilePreservesOtherSettings(t *testing.T) {
	before := []byte("// custom profile\r\nseta r_fullscreen \"1\" // display\r\nseta r_mode \"1920x1080\"\r\nseta com_maxfps \"333\"\r\nbind F5 \"+actionslot 5\"\r\nseta name \"Player\"\r\n")
	cfg := Config{DisplayMode: "windowed", Resolution: "1280x720"}
	if _, err := patchVideoProfile([]byte("/*\nseta r_fullscreen \"1\"\n*/\n"), cfg); err == nil {
		t.Fatal("Block comments must not be edited")
	}
	got, err := patchVideoProfile(before, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Replace(before, []byte("r_fullscreen \"1\""), []byte("r_fullscreen \"0\""), 1)
	want = bytes.Replace(want, []byte("r_mode \"1920x1080\""), []byte("r_mode \"1280x720\""), 1)
	if !bytes.Equal(got, want) {
		t.Fatalf("Unrelated data changed: %q", got)
	}
	again, err := patchVideoProfile(got, cfg)
	if err != nil || !bytes.Equal(again, got) {
		t.Fatal("Not repeatable", err)
	}
	unchanged, err := patchVideoProfile(before, Config{})
	if err != nil || !bytes.Equal(unchanged, before) {
		t.Fatal("Default modified profile")
	}
	chained := []byte("seta r_fullscreen \"1\"; bind X \"say hello\"\n")
	patched, err := patchVideoProfile(chained, cfg)
	if err != nil || !bytes.HasPrefix(patched, chained) || !bytes.Contains(patched, []byte("seta r_fullscreen \"0\"")) {
		t.Fatal("Custom chain was changed", err)
	}
	utf8 := append([]byte{0xef, 0xbb, 0xbf}, before...)
	got, err = patchVideoProfile(utf8, cfg)
	if err != nil || !bytes.Equal(got, append([]byte{0xef, 0xbb, 0xbf}, want...)) {
		t.Fatal("UTF-8 BOM lost", err)
	}
	_, encode, err := decodeProfile([]byte{0xff, 0xfe})
	if err != nil {
		t.Fatal(err)
	}
	utf16 := encode(string(before))
	got, err = patchVideoProfile(utf16, cfg)
	if err != nil || !bytes.Equal(got, encode(string(want))) {
		t.Fatal("UTF-16 data changed", err)
	}
}
func TestVideoProfileBackupDiscoveryAndFailure(t *testing.T) {
	dir, state := t.TempDir(), t.TempDir()
	path := filepath.Join(dir, "storage", "iw5", "players", "config_mp.cfg")
	original := []byte("seta r_fullscreen \"1\"\nbind F6 \"+actionslot 6\"\n")
	if err := atomicWrite(path, original); err != nil {
		t.Fatal(err)
	}
	cfg := Config{PlutoDir: dir, GameDir: t.TempDir(), DisplayMode: "windowed"}
	selected, changed, err := prepareVideoProfile(cfg, state)
	if err != nil || !changed || selected != path {
		t.Fatal(selected, changed, err)
	}
	backups, err := filepath.Glob(filepath.Join(state, "backups", "video-profile", "*.cfg"))
	if err != nil || len(backups) != 1 {
		t.Fatal(backups, err)
	}
	saved, err := os.ReadFile(backups[0])
	if err != nil || !bytes.Equal(saved, original) {
		t.Fatal("Original not preserved", err)
	}
	_, changed, err = prepareVideoProfile(cfg, state)
	if err != nil || changed {
		t.Fatal("Unnecessary second write", err)
	}
	missing := Config{PlutoDir: t.TempDir(), GameDir: t.TempDir(), DisplayMode: "windowed"}
	if _, _, err = prepareVideoProfile(missing, state); err == nil || !strings.Contains(err.Error(), "Geen bestaande") {
		t.Fatal("Missing profile accepted", err)
	}
	if _, changed, err = prepareVideoProfile(Config{}, state); err != nil || changed {
		t.Fatal("Default requires profile", err)
	}
	// Existing backup corruption must stop the update, not overwrite the backup.
	if err = atomicWrite(path, original); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(original)
	backup := filepath.Join(state, "backups", "video-profile", fmt.Sprintf("%x.cfg", hash))
	if err = atomicWrite(backup, []byte("damaged backup")); err != nil {
		t.Fatal(err)
	}
	if _, _, err = prepareVideoProfile(cfg, state); err == nil {
		t.Fatal("Corrupt backup accepted")
	}
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data, original) {
		t.Fatal("Profile changed after backup failure")
	}
}
func TestVideoProfileReadonlyAndFallback(t *testing.T) {
	base, state := t.TempDir(), t.TempDir()
	path := filepath.Join(base, "players2", "config_mp.cfg")
	if err := atomicWrite(path, []byte("seta r_fullscreen \"1\"\n")); err != nil {
		t.Fatal(err)
	}
	cfg := Config{PlutoDir: t.TempDir(), GameDir: base, DisplayMode: "windowed"}
	if err := os.Chmod(path, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0600)
	if _, _, err := prepareVideoProfile(cfg, state); err == nil {
		t.Fatal("Readonly profile modified")
	}
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	selected, changed, err := prepareVideoProfile(cfg, state)
	if err != nil || !changed || selected != path {
		t.Fatal("Game profile fallback failed", selected, err)
	}
}
