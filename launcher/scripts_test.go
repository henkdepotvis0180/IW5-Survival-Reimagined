package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestDeployedProjectScriptsWinArchiveSearch(t *testing.T) {
	payload, err := os.ReadFile("payload.zip")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if _, err = install(payload, root, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	a, err := os.ReadFile(filepath.Join(root, "zz_survival_reimagined.iwd"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "mods/survival/zz_survival_reimagined.iwd"))
	if err != nil || !bytes.Equal(a, b) {
		t.Fatal("Root and mod script patches differ", err)
	}
	patch, err := zip.NewReader(bytes.NewReader(a), int64(len(a)))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range patch.File {
		if entry.Name != strings.ToLower(entry.Name) {
			t.Fatal("Script lookup path is not normalized", entry.Name)
		}
	}
	archives, err := filepath.Glob(filepath.Join(root, "*.iwd"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(archives)))
	for _, name := range []string{"lethalbeats/survival/survivorhandler.gsc", "lethalbeats/survival/dev/testmenu.gsc", "lethalbeats/survival/difficulty.gsc", "lethalbeats/survival/utility.gsc", "maps/mp/gametypes/survival.gsc"} {
		found := false
		for _, path := range archives {
			z, err := zip.OpenReader(path)
			if err != nil {
				t.Fatal(err)
			}
			var selected *zip.File
			for _, f := range z.File {
				if strings.EqualFold(f.Name, name) {
					selected = f
					break
				}
			}
			if selected == nil {
				z.Close()
				continue
			}
			if filepath.Base(path) != "zz_survival_reimagined.iwd" {
				t.Fatal("Original archive shadows project script", name)
			}
			f, err := selected.Open()
			if err != nil {
				t.Fatal(err)
			}
			actual := new(bytes.Buffer)
			_, err = actual.ReadFrom(f)
			f.Close()
			z.Close()
			if err != nil {
				t.Fatal(err)
			}
			// Resolve original casing by walking the source tree.
			var source []byte
			err = filepath.WalkDir("../iw5", func(p string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.IsDir() {
					return nil
				}
				rel, e := filepath.Rel("../iw5", p)
				if e != nil {
					return e
				}
				if strings.EqualFold(filepath.ToSlash(rel), name) {
					source, e = os.ReadFile(p)
					return e
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if source == nil || !bytes.Equal(actual.Bytes(), source) {
				t.Fatal("Deployed script differs from project source", name)
			}
			found = true
			break
		}
		if !found {
			t.Fatal("Project script not deployed", name)
		}
	}
}
