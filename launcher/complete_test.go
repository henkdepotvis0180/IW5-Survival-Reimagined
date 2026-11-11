package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestCompleteOfficialBaseline(t *testing.T) {
	raw, err := os.ReadFile("official-files.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]string
	if err = json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile("payload.zip")
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatal(err)
	}
	for name, checksum := range expected {
		f, err := z.Open("iw5/" + name)
		if err != nil {
			t.Fatal(name, err)
		}
		h := sha256.New()
		_, err = io.Copy(h, f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", h.Sum(nil)) != checksum {
			t.Fatal("Wrong official baseline", name)
		}
	}
	f, err := z.Open("iw5/mods/survival/mod.iwd")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	mod, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"images/lb_box_png.iwi", "images/preview_mp_dlc.iwi"} {
		f, err := mod.Open(name)
		if err != nil {
			t.Fatal(name, err)
		}
		f.Close()
	}
}
