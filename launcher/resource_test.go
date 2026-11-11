package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBundledImageRepair(t *testing.T) {
	payload, err := os.ReadFile("payload.zip")
	if os.IsNotExist(err) {
		t.Skip("Run build.py to generate the distribution payload first")
	}
	if err != nil {
		t.Fatal(err)
	}
	root, backups := t.TempDir(), t.TempDir()
	mod := filepath.Join(root, "mods", "survival", "mod.iwd")
	if err := atomicWrite(mod, []byte("old mod archive")); err != nil {
		t.Fatal(err)
	}
	if _, err := install(payload, root, backups); err != nil {
		t.Fatal(err)
	}
	repair, err := os.ReadFile(filepath.Join(root, "mods", "survival", "survival_ui_assets.iwd"))
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(repair), int64(len(repair)))
	if err != nil {
		t.Fatal(err)
	}
	f, err := z.Open("images/lb_box_png.iwi")
	if err != nil {
		t.Fatal(err)
	}
	image, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(image)) != "b48b0f2301616c22e96c0788f2699ee9c8016491144fef80529db6671a4ad362" {
		t.Fatal("Bundled repair does not contain the original upstream image")
	}
	count, err := install(payload, root, backups)
	if err != nil || count != 0 {
		t.Fatalf("Repair reinstall: %d files, %v", count, err)
	}
	oldHash := sha256.Sum256([]byte("old mod archive"))
	backup, err := os.ReadFile(filepath.Join(backups, "mods", "survival", fmt.Sprintf("mod.iwd.%x.bak", oldHash)))
	if err != nil || string(backup) != "old mod archive" {
		t.Fatalf("Existing mod backup missing: %v", err)
	}
}
