package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVideoOverridesAndPersistence(t *testing.T) {
	args, err := videoArgs(Config{})
	if err != nil || len(args) != 0 {
		t.Fatal("Legacy settings must not override video", args, err)
	}
	cfg := Config{Map: "mp_dome", DisplayMode: "windowed", Resolution: "1280x720", VSync: "off", FPSLimit: 144}
	args, err = videoArgs(cfg)
	want := []string{"+set", "r_fullscreen", "0", "+set", "r_mode", "1280x720", "+set", "r_vsync", "0", "+set", "com_maxfps", "144"}
	if err != nil || !reflect.DeepEqual(args, want) {
		t.Fatal(args, err)
	}
	all, err := launchArgs(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(all[5:5+len(want)], want) || all[len(all)-2] != "+map" {
		t.Fatal("Video options must precede map start", all)
	}
	full, err := videoArgs(Config{DisplayMode: "fullscreen", VSync: "on"})
	if err != nil || !reflect.DeepEqual(full, []string{"+set", "r_fullscreen", "1", "+set", "r_vsync", "1"}) {
		t.Fatal(full, err)
	}
	for _, bad := range []Config{{DisplayMode: "invalid"}, {Resolution: "1280x720;quit"}, {VSync: "invalid"}, {FPSLimit: -1}} {
		if _, err := videoArgs(bad); err == nil {
			t.Fatal("Accepted invalid video override", bad)
		}
	}
	path := filepath.Join(t.TempDir(), "launcher.json")
	if err := saveConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved Config
	if err := json.Unmarshal(data, &saved); err != nil || saved != cfg {
		t.Fatal("Video settings not preserved", saved, err)
	}
}
