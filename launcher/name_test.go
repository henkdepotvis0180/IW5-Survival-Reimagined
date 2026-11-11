package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLANNameIsExplicitAndPreserved(t *testing.T) {
	name := "abcdefghijklmnop" // Sixteen characters, including the final character.
	cfg := Config{GameDir: "game", Map: "mp_dome", PlayerName: name}
	args, err := launchArgs(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if args[2] != "-lan" || args[3] != "+name" || args[4] != name {
		t.Fatalf("LAN name lost or truncated: %v", args)
	}
	path := filepath.Join(t.TempDir(), "launcher.json")
	if err = saveConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved Config
	if err = json.Unmarshal(data, &saved); err != nil || saved.PlayerName != name {
		t.Fatalf("Saved name mismatch: %v", err)
	}
	args, err = launchArgs(Config{Map: "mp_dome"})
	if err != nil || args[4] != "Survivor" {
		t.Fatal("Legacy config has no valid LAN name", args, err)
	}
	for _, bad := range []string{"bad;quit", "bad\nname", "abcdefghijklmnopq"} {
		if _, err = launchArgs(Config{Map: "mp_dome", PlayerName: bad}); err == nil {
			t.Fatalf("Accepted invalid name: %q", bad)
		}
	}
}
