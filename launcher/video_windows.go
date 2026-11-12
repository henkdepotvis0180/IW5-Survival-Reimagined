package main

import "unsafe"

var displayValues = []string{"", "windowed", "fullscreen"}
var resolutionValues = []string{"", "1280x720", "1366x768", "1600x900", "1920x1080", "2560x1440", "3840x2160"}
var vsyncValues = []string{"", "on", "off"}
var fpsValues = []int{0, 60, 120, 144, 165, 240}

func comboIndex(h uintptr, count int) int {
	index, _, _ := sendMessage.Call(h, 0x147, 0, 0)
	if index >= uintptr(count) {
		return 0
	}
	return int(index)
}
func comboValue(h uintptr, values []string) string { return values[comboIndex(h, len(values))] }
func (a *launcherUI) combo(labels, values []string, current string, x, y int32, id uintptr) uintptr {
	h := a.control("COMBOBOX", "", 0x210003, x, y, 320, 230, id)
	selected := 0
	for i, label := range labels {
		sendMessage.Call(h, 0x143, 0, uintptr(unsafe.Pointer(wide(label))))
		if values[i] == current {
			selected = i
		}
	}
	sendMessage.Call(h, 0x14E, uintptr(selected), 0)
	a.videoControls = append(a.videoControls, h)
	a.controls = append(a.controls, h)
	return h
}
func (a *launcherUI) initVideoControls(cfg Config) {
	label := func(value string, x, y int32) {
		h := a.control("STATIC", value, 0, x, y, 320, 22, 0)
		a.videoControls = append(a.videoControls, h)
	}
	label("SCHERMMODUS", 280, 140)
	label("RESOLUTIE", 640, 140)
	a.displayMode = a.combo([]string{"Bestaande instelling", "Venstermodus", "Volledig scherm"}, displayValues, cfg.DisplayMode, 280, 168, 110)
	a.resolution = a.combo([]string{"Bestaande instelling", "1280 x 720", "1366 x 768", "1600 x 900", "1920 x 1080", "2560 x 1440", "3840 x 2160"}, resolutionValues, cfg.Resolution, 640, 168, 111)
	label("VSYNC", 280, 224)
	label("FPS-LIMIET", 640, 224)
	a.vsync = a.combo([]string{"Bestaande instelling", "Aan", "Uit"}, vsyncValues, cfg.VSync, 280, 252, 112)
	fpsText := []string{"Bestaande instelling", "60 FPS", "120 FPS", "144 FPS", "165 FPS", "240 FPS"}
	fpsStrings := []string{"", "60", "120", "144", "165", "240"}
	fpsCurrent := ""
	for i, v := range fpsValues {
		if v == cfg.FPSLimit {
			fpsCurrent = fpsStrings[i]
		}
	}
	a.fps = a.combo(fpsText, fpsStrings, fpsCurrent, 640, 252, 113)
	note := a.control("STATIC", "Wijzigingen gelden bij de volgende gamestart.\r\nDe launcher maakt eerst een backup van je bestaande videoprofiel.\r\nKies een resolutie die je scherm ondersteunt.", 0, 280, 326, 680, 56, 0)
	preset := a.control("BUTTON", "VENSTER VOOR SCREENSHOTS", 0x1000B, 280, 400, 400, 48, 114)
	save := a.control("BUTTON", "VIDEO OPSLAAN", 0x1000B, 280, 472, 320, 48, idSave)
	hint := a.control("STATIC", "Screenshot-preset: venstermodus op 1280 x 720.\r\nOpslaan en daarna Start Survival om de keuze toe te passen.", 0, 280, 526, 680, 36, 0)
	a.videoControls = append(a.videoControls, note, preset, save, hint)
	a.controls = append(a.controls, preset, save)
}
func (a *launcherUI) showVideo() {
	a.showSettings(false)
	a.video = true
	for _, h := range a.playControls {
		user32.NewProc("ShowWindow").Call(h, 0)
	}
	for _, h := range a.videoControls {
		user32.NewProc("ShowWindow").Call(h, 5)
	}
	user32.NewProc("RedrawWindow").Call(a.hwnd, 0, 0, 0x85)
}
