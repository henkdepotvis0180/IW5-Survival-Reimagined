package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

//go:embed payload.zip
var payload []byte

func message(text string) {
	title, _ := syscall.UTF16PtrFromString("Survival Reimagined")
	body, _ := syscall.UTF16PtrFromString(text)
	syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x10)
}

func pickFolder(title string) (string, error) {
	// Windows PowerShell and WinForms ship with supported Windows installations.
	// Fixed titles only; no user-controlled content is interpolated into this script.
	script := `Add-Type -AssemblyName System.Windows.Forms; $d=New-Object System.Windows.Forms.FolderBrowserDialog; $d.Description='` + title + `'; $d.ShowNewFolderButton=$false; if($d.ShowDialog() -eq 'OK'){[Console]::OutputEncoding=[System.Text.Encoding]::UTF8; [Console]::Write($d.SelectedPath)}`
	units := utf16.Encode([]rune(script))
	b := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	ps := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	cmd := exec.Command(ps, "-NoProfile", "-STA", "-EncodedCommand", base64.StdEncoding.EncodeToString(b))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Mapkeuze mislukt: %w", err)
	}
	path := strings.TrimPrefix(strings.TrimSpace(string(out)), "\ufeff")
	if path == "" {
		return "", fmt.Errorf("Geannuleerd; er is niets gestart")
	}
	return path, nil
}

func run() error {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return fmt.Errorf("LOCALAPPDATA ontbreekt")
	}
	state := filepath.Join(local, "SurvivalReimagined")
	cfgPath := filepath.Join(state, "launcher.json")
	cfg := Config{PlutoDir: filepath.Join(local, "Plutonium"), Map: "mp_dome"}
	if data, err := os.ReadFile(cfgPath); err == nil {
		if err = json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("Ongeldige instellingen in %s: %w", cfgPath, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if !gameValid(cfg.GameDir) {
		folder, err := pickFolder("Kies je MW3-map met main en zone/english")
		if err != nil {
			return err
		}
		if !gameValid(folder) {
			return fmt.Errorf("Geen volledige Engelse MW3-installatie gevonden in %s (main/iw_00.iwd en zone/english vereist)", folder)
		}
		cfg.GameDir = folder
	}
	bootstrap := filepath.Join(cfg.PlutoDir, "bin", "plutonium-bootstrapper-win32.exe")
	if !fileExists(bootstrap) {
		folder, err := pickFolder("Kies je Plutonium-map met de submap bin")
		if err != nil {
			return err
		}
		cfg.PlutoDir = folder
		bootstrap = filepath.Join(folder, "bin", "plutonium-bootstrapper-win32.exe")
		if !fileExists(bootstrap) {
			return fmt.Errorf("Plutonium ontbreekt. Installeer/update en start eerst de officiele Plutonium-launcher")
		}
	}
	args, err := launchArgs(cfg)
	if err != nil {
		return err
	}
	if err = saveConfig(cfgPath, cfg); err != nil {
		return err
	}
	_, err = install(payload, filepath.Join(cfg.PlutoDir, "storage", "iw5"), filepath.Join(state, "backups"))
	if err != nil {
		return err
	}
	log := filepath.Join(state, "last-launch.log")
	if err = os.MkdirAll(state, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(log, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, "Starting Survival Reimagined in LAN mode on", cfg.Map)
	cmd := exec.Command(bootstrap, args...)
	cmd.Dir = cfg.PlutoDir
	cmd.Stdout = f
	cmd.Stderr = f
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Plutonium is gestopt met een fout: %w. Zie %s", err, log)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		message(err.Error())
	}
}
