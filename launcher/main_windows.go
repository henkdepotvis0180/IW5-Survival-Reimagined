package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

//go:embed payload.zip
var payload []byte

var user32 = syscall.NewLazyDLL("user32.dll")
var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var sendMessage = user32.NewProc("SendMessageW")
var setText = user32.NewProc("SetWindowTextW")
var application *launcherUI

func wide(text string) *uint16 { p, _ := syscall.UTF16PtrFromString(text); return p }

func message(owner uintptr, text string) {
	user32.NewProc("MessageBoxW").Call(owner, uintptr(unsafe.Pointer(wide(text))), uintptr(unsafe.Pointer(wide("Survival Reimagined"))), 0x10)
}

// Native Windows folder chooser: no PowerShell subprocess or encoded scripts.
func pickFolder(owner uintptr, title string) (string, error) {
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
	info := browseInfo{owner: owner, displayName: &display[0], title: caption, flags: 0x41}
	item, _, _ := shell.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&info)))
	if item == 0 {
		return "", nil
	}
	defer ole.NewProc("CoTaskMemFree").Call(item)
	var path [260]uint16
	ok, _, _ := shell.NewProc("SHGetPathFromIDListW").Call(item, uintptr(unsafe.Pointer(&path[0])))
	if ok == 0 {
		return "", fmt.Errorf("Kies een gewone lokale map")
	}
	return syscall.UTF16ToString(path[:]), nil
}

const (
	idGameBrowse  = 101
	idPlutoBrowse = 102
	idSave        = 103
	idStart       = 104
	idLog         = 105
	wmUpdate      = 0x8001
)

type uiEvent struct {
	text string
	done bool
	err  error
}
type launcherUI struct {
	hwnd, game, pluto, mapName, test, status uintptr
	controls                                 []uintptr
	state                                    string
	busy                                     bool
	events                                   chan uiEvent
}

func (a *launcherUI) control(class, text string, style uint32, x, y, w, h int32, id uintptr) uintptr {
	handle, _, _ := user32.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(wide(class))), uintptr(unsafe.Pointer(wide(text))), uintptr(style|0x50000000), uintptr(x), uintptr(y), uintptr(w), uintptr(h), a.hwnd, id, 0, 0)
	if handle == 0 {
		panic("Windows-control maken mislukt")
	}
	font, _, _ := syscall.NewLazyDLL("gdi32.dll").NewProc("GetStockObject").Call(17)
	sendMessage.Call(handle, 0x30, font, 1)
	return handle
}

func (a *launcherUI) initControls(cfg Config) {
	a.control("STATIC", "SURVIVAL REIMAGINED", 0, 24, 20, 650, 24, 0)
	a.control("STATIC", "Kies je bestaande installaties en start een Survival-match in LAN-modus.", 0, 24, 50, 650, 22, 0)
	a.control("STATIC", "MW3-installatiemap", 0, 24, 90, 650, 20, 0)
	a.game = a.control("EDIT", cfg.GameDir, 0x00810080, 24, 114, 540, 28, 0)
	gameBrowse := a.control("BUTTON", "Bladeren...", 0x10000, 576, 114, 112, 28, idGameBrowse)
	a.control("STATIC", "Plutonium-map (met de submap bin)", 0, 24, 156, 650, 20, 0)
	a.pluto = a.control("EDIT", cfg.PlutoDir, 0x00810080, 24, 180, 540, 28, 0)
	plutoBrowse := a.control("BUTTON", "Bladeren...", 0x10000, 576, 180, 112, 28, idPlutoBrowse)
	a.control("STATIC", "Startmap", 0, 24, 224, 120, 20, 0)
	a.mapName = a.control("EDIT", cfg.Map, 0x00810080, 24, 248, 200, 28, 0)
	a.test = a.control("BUTTON", "Host-testmenu met geld inschakelen", 0x10003, 246, 248, 442, 28, 0)
	if cfg.TestMenu {
		sendMessage.Call(a.test, 0xF1, 1, 0)
	}
	a.control("STATIC", "De mod wordt bijgewerkt in Plutonium-storage. Vervangen bestanden krijgen een backup.\r\nSluit de game voordat je start. Deze testversie is niet digitaal ondertekend.", 0, 24, 296, 664, 44, 0)
	save := a.control("BUTTON", "Instellingen opslaan", 0x10000, 24, 354, 172, 34, idSave)
	start := a.control("BUTTON", "Start Survival", 0x10001, 208, 354, 172, 34, idStart)
	a.control("BUTTON", "Log bekijken", 0x10000, 516, 354, 172, 34, idLog)
	a.status = a.control("STATIC", "Gereed. Je kunt de installatiemappen hier altijd wijzigen.", 0, 24, 408, 664, 42, 0)
	a.controls = []uintptr{a.game, a.pluto, a.mapName, a.test, gameBrowse, plutoBrowse, save, start}
}

func windowText(hwnd uintptr) string {
	length, _, _ := user32.NewProc("GetWindowTextLengthW").Call(hwnd)
	buf := make([]uint16, length+1)
	user32.NewProc("GetWindowTextW").Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func (a *launcherUI) config() (Config, error) {
	checked, _, _ := sendMessage.Call(a.test, 0xF0, 0, 0)
	cfg := Config{GameDir: windowText(a.game), PlutoDir: windowText(a.pluto), Map: windowText(a.mapName), TestMenu: checked == 1}
	if !gameValid(cfg.GameDir) {
		return cfg, fmt.Errorf("Kies een geldige MW3-map met main/iw_00.iwd en zone/english")
	}
	if !fileExists(filepath.Join(cfg.PlutoDir, "bin", "plutonium-bootstrapper-win32.exe")) {
		return cfg, fmt.Errorf("Kies de Plutonium-map met bin/plutonium-bootstrapper-win32.exe")
	}
	_, err := launchArgs(cfg)
	return cfg, err
}

func (a *launcherUI) statusText(text string) {
	setText.Call(a.status, uintptr(unsafe.Pointer(wide(text))))
}
func (a *launcherUI) enable(enabled bool) {
	value := uintptr(0)
	if enabled {
		value = 1
	}
	for _, control := range a.controls {
		user32.NewProc("EnableWindow").Call(control, value)
	}
}
func (a *launcherUI) post(event uiEvent) {
	a.events <- event
	user32.NewProc("PostMessageW").Call(a.hwnd, wmUpdate, 0, 0)
}

func (a *launcherUI) command(id uintptr) {
	if id == idLog {
		file := filepath.Join(a.state, "last-launch.log")
		notepad := filepath.Join(os.Getenv("SystemRoot"), "System32", "notepad.exe")
		if err := exec.Command(notepad, file).Start(); err != nil {
			message(a.hwnd, err.Error())
		}
		return
	}
	if a.busy {
		return
	}
	if id == idGameBrowse || id == idPlutoBrowse {
		title := "Kies je MW3-installatiemap"
		target := a.game
		if id == idPlutoBrowse {
			title = "Kies je Plutonium-map met de submap bin"
			target = a.pluto
		}
		path, err := pickFolder(a.hwnd, title)
		if err != nil {
			message(a.hwnd, err.Error())
			return
		}
		if path != "" {
			setText.Call(target, uintptr(unsafe.Pointer(wide(path))))
		}
		return
	}
	if id != idSave && id != idStart {
		return
	}
	cfg, err := a.config()
	if err != nil {
		message(a.hwnd, err.Error())
		return
	}
	if err = saveConfig(filepath.Join(a.state, "launcher.json"), cfg); err != nil {
		message(a.hwnd, err.Error())
		return
	}
	if id == idSave {
		a.statusText("Instellingen opgeslagen.")
		return
	}
	a.busy = true
	a.enable(false)
	a.statusText("Modbestanden controleren en backups maken...")
	go func() {
		defer func() {
			if v := recover(); v != nil {
				a.post(uiEvent{done: true, err: fmt.Errorf("Onverwachte fout: %v", v)})
			}
		}()
		if err := launch(cfg, a.state, a.post); err != nil {
			a.post(uiEvent{done: true, err: err})
			return
		}
		a.post(uiEvent{done: true, text: "Plutonium is afgesloten. Je kunt opnieuw starten of je mappen wijzigen."})
	}()
}

func launch(cfg Config, state string, post func(uiEvent)) error {
	log.Println("Modbestanden controleren...")
	count, err := install(payload, filepath.Join(cfg.PlutoDir, "storage", "iw5"), filepath.Join(state, "backups"))
	if err != nil {
		return err
	}
	log.Printf("%d modbestanden bijgewerkt", count)
	args, err := launchArgs(cfg)
	if err != nil {
		return err
	}
	cmd := exec.Command(filepath.Join(cfg.PlutoDir, "bin", "plutonium-bootstrapper-win32.exe"), args...)
	cmd.Dir = cfg.PlutoDir
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("Plutonium starten mislukt: %w", err)
	}
	post(uiEvent{text: "Plutonium is gestart. Wacht tot de game opent; zie het log als er niets verschijnt."})
	log.Println("Plutonium gestart op", cfg.Map, "in LAN-modus")
	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("Plutonium is gestopt met een fout: %w. Zie Log bekijken", err)
	}
	return nil
}

func windowProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	a := application
	switch msg {
	case 0x111:
		if a != nil {
			a.command(wparam & 0xffff)
		}
		return 0
	case wmUpdate:
		select {
		case event := <-a.events:
			if event.done {
				a.busy = false
				a.enable(true)
			}
			if event.err != nil {
				log.Println("FOUT:", event.err)
				a.statusText("Starten mislukt. Zie Log bekijken voor details.")
				message(hwnd, event.err.Error())
			} else {
				a.statusText(event.text)
			}
		default:
		}
		return 0
	case 0x10:
		if a != nil && a.busy {
			message(hwnd, "Sluit eerst de game. Tijdens installeren en starten blijft de launcher geopend.")
			return 0
		}
		user32.NewProc("DestroyWindow").Call(hwnd)
		return 0
	case 0x2:
		user32.NewProc("PostQuitMessage").Call(0)
		return 0
	}
	result, _, _ := user32.NewProc("DefWindowProcW").Call(hwnd, uintptr(msg), wparam, lparam)
	return result
}

func runUI(state string) error {
	cfg := Config{PlutoDir: filepath.Join(os.Getenv("LOCALAPPDATA"), "Plutonium"), Map: "mp_dome"}
	if data, err := os.ReadFile(filepath.Join(state, "launcher.json")); err == nil {
		var saved Config
		if err = json.Unmarshal(data, &saved); err != nil {
			message(0, "Opgeslagen instellingen zijn ongeldig. Kies je mappen opnieuw en sla ze op.")
		} else {
			cfg = saved
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	app := &launcherUI{state: state, events: make(chan uiEvent, 8)}
	application = app
	instance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	icon, _, _ := user32.NewProc("LoadIconW").Call(0, 32512)
	type windowClass struct {
		size, style                        uint32
		callback                           uintptr
		classExtra, windowExtra            int32
		instance, icon, cursor, background uintptr
		menuName, className                *uint16
		smallIcon                          uintptr
	}
	class := windowClass{callback: syscall.NewCallback(windowProc), instance: instance, icon: icon, cursor: cursor, background: 16, className: wide("SurvivalReimaginedLauncher"), smallIcon: icon}
	class.size = uint32(unsafe.Sizeof(class))
	atom, _, err := user32.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&class)))
	if atom == 0 {
		return fmt.Errorf("Venster registreren mislukt: %v", err)
	}
	screenW, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	screenH, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	x, y := uintptr(0), uintptr(0)
	if screenW > 736 {
		x = (screenW - 736) / 2
	}
	if screenH > 510 {
		y = (screenH - 510) / 2
	}
	hwnd, _, err := user32.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(class.className)), uintptr(unsafe.Pointer(wide("Survival Reimagined Launcher 0.2.0 - testversie"))), 0x00CA0000, x, y, 736, 510, 0, 0, instance, 0)
	if hwnd == 0 {
		return fmt.Errorf("Venster openen mislukt: %v", err)
	}
	app.hwnd = hwnd
	app.initControls(cfg)
	user32.NewProc("ShowWindow").Call(hwnd, 1)
	user32.NewProc("UpdateWindow").Call(hwnd)
	type winMessage struct {
		hwnd           uintptr
		message        uint32
		wparam, lparam uintptr
		time           uint32
		x, y           int32
		private        uint32
	}
	var msg winMessage
	for {
		result, _, err := user32.NewProc("GetMessageW").Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) == -1 {
			return fmt.Errorf("Windows-berichten lezen mislukt: %v", err)
		}
		if result == 0 {
			break
		}
		handled, _, _ := user32.NewProc("IsDialogMessageW").Call(hwnd, uintptr(unsafe.Pointer(&msg)))
		if handled == 0 {
			user32.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&msg)))
			user32.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
	return nil
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer func() {
		if v := recover(); v != nil {
			log.Printf("FOUT: %v", v)
			message(0, fmt.Sprint(v))
		}
	}()
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		message(0, "LOCALAPPDATA ontbreekt")
		return
	}
	state := filepath.Join(local, "SurvivalReimagined")
	if err := os.MkdirAll(state, 0755); err != nil {
		message(0, err.Error())
		return
	}
	file, err := os.OpenFile(filepath.Join(state, "last-launch.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		message(0, "Logbestand openen mislukt: "+err.Error())
		return
	}
	defer file.Close()
	log.SetOutput(file)
	log.Println("Launcher 0.2.0 gestart")
	if err = runUI(state); err != nil {
		log.Println("FOUT:", err)
		message(0, err.Error())
	}
}
