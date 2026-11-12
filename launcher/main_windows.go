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
	"time"
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
	idPlay        = 106
	idSettings    = 107
	idVideo       = 109
	wmUpdate      = 0x8001
)

type uiEvent struct {
	text string
	done bool
	err  error
}
type launcherUI struct {
	hwnd, game, pluto, mapName, playerName, test, status uintptr
	controls                                             []uintptr
	playControls, settingControls                        []uintptr
	settings                                             bool
	video                                                bool
	videoControls                                        []uintptr
	displayMode, resolution, vsync, fps                  uintptr
	state                                                string
	busy                                                 bool
	events                                               chan uiEvent
}

func (a *launcherUI) control(class, text string, style uint32, x, y, w, h int32, id uintptr) uintptr {
	handle, _, _ := user32.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(wide(class))), uintptr(unsafe.Pointer(wide(text))), uintptr(style|0x50000000), uintptr(x), uintptr(y), uintptr(w), uintptr(h), a.hwnd, id, 0, 0)
	if handle == 0 {
		panic("Windows-control maken mislukt")
	}
	font := theme.body
	sendMessage.Call(handle, 0x30, font, 1)
	return handle
}

func (a *launcherUI) initControls(cfg Config) {
	play := a.control("BUTTON", "01   SPELEN", 0x1000B, 24, 180, 184, 48, idPlay)
	settings := a.control("BUTTON", "02   INSTELLINGEN", 0x1000B, 24, 238, 184, 48, idSettings)
	videoTab := a.control("BUTTON", "03   VIDEO", 0x1000B, 24, 296, 184, 48, idVideo)
	a.control("BUTTON", "LOG BEKIJKEN", 0x1000B, 24, 556, 184, 40, idLog)
	start := a.control("BUTTON", "START SURVIVAL   >", 0x1000B, 280, 456, 440, 64, idStart)
	a.playControls = []uintptr{start}
	label := func(text string, y int32) {
		h := a.control("STATIC", text, 0, 280, y, 680, 22, 0)
		a.settingControls = append(a.settingControls, h)
	}
	label("MW3-INSTALLATIEMAP", 140)
	a.game = a.control("EDIT", cfg.GameDir, 0x00810080, 280, 168, 548, 32, 0)
	gameBrowse := a.control("BUTTON", "BLADEREN", 0x1000B, 844, 168, 116, 32, idGameBrowse)
	label("PLUTONIUM-MAP", 224)
	a.pluto = a.control("EDIT", cfg.PlutoDir, 0x00810080, 280, 252, 548, 32, 0)
	plutoBrowse := a.control("BUTTON", "BLADEREN", 0x1000B, 844, 252, 116, 32, idPlutoBrowse)
	mapLabel := a.control("STATIC", "STARTMAP", 0, 280, 308, 240, 22, 0)
	nameLabel := a.control("STATIC", "SPELERNAAM (LAN)", 0, 560, 308, 400, 22, 0)
	a.settingControls = append(a.settingControls, mapLabel, nameLabel)
	a.mapName = a.control("EDIT", cfg.Map, 0x00810080, 280, 336, 240, 32, 0)
	a.playerName = a.control("EDIT", cfg.PlayerName, 0x00810080, 560, 336, 400, 32, 108)
	sendMessage.Call(a.playerName, 0xC5, 16, 0)
	a.test = a.control("BUTTON", "Geldtestmenu inschakelen (actieslot 5)", 0x10003, 280, 406, 680, 32, 0)
	syscall.NewLazyDLL("uxtheme.dll").NewProc("SetWindowTheme").Call(a.test, uintptr(unsafe.Pointer(wide(""))), uintptr(unsafe.Pointer(wide(""))))
	if cfg.TestMenu {
		sendMessage.Call(a.test, 0xF1, 1, 0)
	}
	save := a.control("BUTTON", "INSTELLINGEN OPSLAAN", 0x1000B, 280, 472, 320, 48, idSave)
	hint := a.control("STATIC", "Vervangen modbestanden krijgen automatisch een backup.\r\nSluit de game voordat je de installatie bijwerkt.", 0, 280, 526, 680, 36, 0)
	a.settingControls = append(a.settingControls, a.game, a.pluto, a.mapName, a.playerName, a.test, gameBrowse, plutoBrowse, save, hint)
	a.status = a.control("STATIC", "Gereed om te starten.", 0, 280, 610, 680, 42, 0)
	a.controls = []uintptr{a.game, a.pluto, a.mapName, a.playerName, a.test, gameBrowse, plutoBrowse, save, start, play, settings}
	a.initVideoControls(cfg)
	a.controls = append(a.controls, videoTab)
	a.showSettings(cfg.GameDir == "")
}

func windowText(hwnd uintptr) string {
	length, _, _ := user32.NewProc("GetWindowTextLengthW").Call(hwnd)
	buf := make([]uint16, length+1)
	user32.NewProc("GetWindowTextW").Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func (a *launcherUI) config() (Config, error) {
	checked, _, _ := sendMessage.Call(a.test, 0xF0, 0, 0)
	cfg := Config{GameDir: windowText(a.game), PlutoDir: windowText(a.pluto), Map: windowText(a.mapName), PlayerName: windowText(a.playerName), TestMenu: checked == 1, DisplayMode: comboValue(a.displayMode, displayValues), Resolution: comboValue(a.resolution, resolutionValues), VSync: comboValue(a.vsync, vsyncValues), FPSLimit: fpsValues[comboIndex(a.fps, len(fpsValues))]}
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
	user32.NewProc("InvalidateRect").Call(a.hwnd, 0, 0)
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
	if id == idVideo {
		if !a.busy {
			a.showVideo()
		}
		return
	}
	if id == idPlay || id == idSettings {
		if !a.busy {
			a.showSettings(id == idSettings)
		}
		return
	}
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
	if id == 114 {
		sendMessage.Call(a.displayMode, 0x14E, 1, 0)
		sendMessage.Call(a.resolution, 0x14E, 1, 0)
		a.statusText("Screenshot-preset gekozen. Sla op en start de game opnieuw.")
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
		a.showSettings(true)
		message(a.hwnd, err.Error())
		return
	}
	if err = saveConfig(filepath.Join(a.state, "launcher.json"), cfg); err != nil {
		message(a.hwnd, err.Error())
		return
	}
	if id == idSave {
		a.statusText("Instellingen opgeslagen.")
		a.showSettings(false)
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
	startedAt := time.Now()
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("Plutonium starten mislukt: %w", err)
	}
	post(uiEvent{text: "Plutonium is gestart. Wacht tot de game opent; zie het log als er niets verschijnt."})
	log.Printf("Plutonium gestart: PID=%d map=%s LAN-naam=%q", cmd.Process.Pid, cfg.Map, cfg.PlayerName)
	if err = cmd.Wait(); err != nil {
		log.Printf("Plutonium afgesloten: PID=%d exit=%d duur=%s", cmd.Process.Pid, cmd.ProcessState.ExitCode(), time.Since(startedAt).Round(time.Second))
		return fmt.Errorf("Plutonium is gestopt met een fout: %w. Zie Log bekijken", err)
	}
	return nil
}

func windowProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	a := application
	switch msg {
	case 0xF:
		paintTheme(hwnd)
		return 0
	case 0x14:
		return 1
	case 0x2B:
		drawThemedButton(lparam)
		return 1
	case 0x133, 0x134, 0x135, 0x138:
		gdi.NewProc("SetTextColor").Call(wparam, uintptr(colorText))
		gdi.NewProc("SetBkColor").Call(wparam, uintptr(colorPanel))
		return theme.panelBrush
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
	initTheme()
	defer closeTheme()
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
	if cfg.PlayerName == "" {
		cfg.PlayerName = "Survivor"
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
	class := windowClass{callback: syscall.NewCallback(windowProc), instance: instance, icon: icon, cursor: cursor, background: theme.backgroundBrush, className: wide("SurvivalReimaginedLauncher"), smallIcon: icon}
	class.size = uint32(unsafe.Sizeof(class))
	atom, _, err := user32.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&class)))
	if atom == 0 {
		return fmt.Errorf("Venster registreren mislukt: %v", err)
	}
	screenW, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	screenH, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	x, y := uintptr(0), uintptr(0)
	if screenW > 1040 {
		x = (screenW - 1040) / 2
	}
	if screenH > 720 {
		y = (screenH - 720) / 2
	}
	hwnd, _, err := user32.NewProc("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(class.className)), uintptr(unsafe.Pointer(wide("Survival Reimagined Launcher 0.5.0 - testversie"))), 0x02CA0000, x, y, 1040, 720, 0, 0, instance, 0)
	if hwnd == 0 {
		return fmt.Errorf("Venster openen mislukt: %v", err)
	}
	app.hwnd = hwnd
	dark := int32(1)
	syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute").Call(hwnd, 20, uintptr(unsafe.Pointer(&dark)), 4)
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
	log.Println("Launcher 0.5.0 gestart")
	if err = runUI(state); err != nil {
		log.Println("FOUT:", err)
		message(0, err.Error())
	}
}
