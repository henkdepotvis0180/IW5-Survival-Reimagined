package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

//go:embed payload.zip
var payload []byte

func message(text string) {
	title, _ := syscall.UTF16PtrFromString("Survival Reimagined")
	body, _ := syscall.UTF16PtrFromString(text)
	syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x10)
}

// Yes keeps both locations; No lets the player select both again.
func chooseSavedLocations(cfg Config) (bool, error) {
	title, _ := syscall.UTF16PtrFromString("Survival Reimagined - installatiemappen")
	body, _ := syscall.UTF16PtrFromString("Opgeslagen locaties gebruiken?\n\nMW3: " + cfg.GameDir + "\nPlutonium: " + cfg.PlutoDir + "\n\nJa: starten met deze locaties\nNee: beide mappen opnieuw kiezen\nAnnuleren: stoppen zonder wijzigingen")
	owner, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
	result, _, err := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(owner, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x23)
	switch result {
	case 6:
		return false, nil
	case 7:
		return true, nil
	case 2:
		return false, fmt.Errorf("Geannuleerd; installatiemappen blijven ongewijzigd")
	default:
		return false, fmt.Errorf("Locatiekeuze openen mislukt: %v", err)
	}
}

// Native Windows folder chooser: no PowerShell subprocess or encoded scripts.
func pickFolder(title string) (string, error) {
	log.Println(title)
	ole := syscall.NewLazyDLL("ole32.dll")
	hr, _, _ := ole.NewProc("OleInitialize").Call(0)
	if int32(hr) < 0 {
		return "", fmt.Errorf("Windows-mapkeuze initialiseren mislukt (HRESULT 0x%x)", hr)
	}
	defer ole.NewProc("OleUninitialize").Call()
	shell := syscall.NewLazyDLL("shell32.dll")
	type browseInfo struct {
		owner       uintptr
		root        uintptr
		displayName *uint16
		title       *uint16
		flags       uint32
		callback    uintptr
		param       uintptr
		image       int32
	}
	caption, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return "", err
	}
	var display [260]uint16
	owner, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
	info := browseInfo{owner: owner, displayName: &display[0], title: caption, flags: 0x41}
	item, _, _ := shell.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&info)))
	if item == 0 {
		return "", fmt.Errorf("Mapkeuze geannuleerd; er is niets gestart")
	}
	defer ole.NewProc("CoTaskMemFree").Call(item)
	var path [260]uint16
	ok, _, _ := shell.NewProc("SHGetPathFromIDListW").Call(item, uintptr(unsafe.Pointer(&path[0])))
	if ok == 0 {
		return "", fmt.Errorf("Kies een gewone lokale map")
	}
	return syscall.UTF16ToString(path[:]), nil
}

func run() error {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return fmt.Errorf("LOCALAPPDATA ontbreekt")
	}
	state := filepath.Join(local, "SurvivalReimagined")
	log.Println("Instellingen controleren...")
	cfgPath := filepath.Join(state, "launcher.json")
	cfg := Config{PlutoDir: filepath.Join(local, "Plutonium"), Map: "mp_dome"}
	if data, err := os.ReadFile(cfgPath); err == nil {
		if err = json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("Ongeldige instellingen in %s: %w", cfgPath, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	reselect := false
	if cfg.GameDir != "" {
		var err error
		reselect, err = chooseSavedLocations(cfg)
		if err != nil {
			return err
		}
	}
	if reselect || !gameValid(cfg.GameDir) {
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
	if reselect || !fileExists(bootstrap) {
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
	log.Println("MW3-map:", cfg.GameDir)
	log.Println("Plutonium:", cfg.PlutoDir)
	if err = saveConfig(cfgPath, cfg); err != nil {
		return err
	}
	log.Println("Modbestanden controleren en backups maken...")
	count, err := install(payload, filepath.Join(cfg.PlutoDir, "storage", "iw5"), filepath.Join(state, "backups"))
	if err != nil {
		return err
	}
	log.Printf("%d modbestanden bijgewerkt. Plutonium starten op %s (LAN)...", count, cfg.Map)
	cmd := exec.Command(bootstrap, args...)
	cmd.Dir = cfg.PlutoDir
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	started := time.Now()
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Plutonium is gestopt met een fout: %w. Zie %s", err, filepath.Join(state, "last-launch.log"))
	}
	log.Printf("Plutonium-bootstrapper afgesloten na %s (exitcode 0). Als er geen game verscheen, deel dit logbestand.", time.Since(started).Round(time.Second))
	return nil
}

func main() {
	// Folder dialogs use COM and must run on one OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fmt.Println("Survival Reimagined Launcher v0.1.3 - diagnosevenster")
	log.SetOutput(os.Stdout)
	local := os.Getenv("LOCALAPPDATA")
	if local != "" {
		state := filepath.Join(local, "SurvivalReimagined")
		if err := os.MkdirAll(state, 0755); err == nil {
			file, err := os.OpenFile(filepath.Join(state, "last-launch.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err == nil {
				defer file.Close()
				log.SetOutput(io.MultiWriter(os.Stdout, file))
			} else {
				log.Println("Logbestand openen mislukt:", err)
			}
		} else {
			log.Println("Logmap maken mislukt:", err)
		}
	}
	defer func() {
		if value := recover(); value != nil {
			log.Printf("Onverwachte fout: %v", value)
			message(fmt.Sprint(value))
		}
		fmt.Println("Druk op Enter om dit venster te sluiten.")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}()
	// Explorer, shortcuts and terminals may provide different working directories.
	// Match starting the executable after cd into its own directory.
	executable, err := os.Executable()
	if err != nil {
		log.Println("FOUT: locatie van de launcher bepalen:", err)
		message(err.Error())
		return
	}
	if err = os.Chdir(filepath.Dir(executable)); err != nil {
		log.Println("FOUT: werkmap instellen:", err)
		message(err.Error())
		return
	}
	log.Println("Werkmap launcher:", filepath.Dir(executable))
	log.Println("Launcher gestart. Sluit bestaande MW3-sessies voordat je verdergaat.")
	if err := run(); err != nil {
		log.Println("FOUT:", err)
		message(err.Error())
	}
}
