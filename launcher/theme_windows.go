package main

import (
	"strings"
	"syscall"
	"unsafe"
)

var gdi = syscall.NewLazyDLL("gdi32.dll")

const (
	colorBackground uint32 = 0x0C100D
	colorPanel      uint32 = 0x111711
	colorText       uint32 = 0xDDE6DF
	colorMuted      uint32 = 0x8C9B8F
	colorGreen      uint32 = 0x65C6A9
	colorLine       uint32 = 0x29362B
)

var theme struct{ backgroundBrush, panelBrush, body, small, title, display, subhead uintptr }

type rect struct{ left, top, right, bottom int32 }

func signed(n int32) uintptr { return uintptr(int64(n)) }
func font(height int32, weight uintptr, face string) uintptr {
	h, _, _ := gdi.NewProc("CreateFontW").Call(signed(-height), 0, 0, 0, weight, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wide(face))))
	return h
}
func initTheme() {
	theme.backgroundBrush, _, _ = gdi.NewProc("CreateSolidBrush").Call(uintptr(colorBackground))
	theme.panelBrush, _, _ = gdi.NewProc("CreateSolidBrush").Call(uintptr(colorPanel))
	theme.body = font(15, 400, "Segoe UI")
	theme.small = font(12, 500, "Consolas")
	theme.title = font(24, 600, "Bahnschrift")
	theme.display = font(68, 700, "Bahnschrift SemiCondensed")
	theme.subhead = font(30, 600, "Bahnschrift")
}
func closeTheme() {
	for _, h := range []uintptr{theme.backgroundBrush, theme.panelBrush, theme.body, theme.small, theme.title, theme.display, theme.subhead} {
		gdi.NewProc("DeleteObject").Call(h)
	}
}
func fill(dc uintptr, r rect, color uint32) {
	b, _, _ := gdi.NewProc("CreateSolidBrush").Call(uintptr(color))
	user32.NewProc("FillRect").Call(dc, uintptr(unsafe.Pointer(&r)), b)
	gdi.NewProc("DeleteObject").Call(b)
}
func text(dc uintptr, value string, r rect, color uint32, f uintptr, flags uintptr) {
	old, _, _ := gdi.NewProc("SelectObject").Call(dc, f)
	defer gdi.NewProc("SelectObject").Call(dc, old)
	gdi.NewProc("SetBkMode").Call(dc, 1)
	gdi.NewProc("SetTextColor").Call(dc, uintptr(color))
	units := syscall.StringToUTF16(value)
	user32.NewProc("DrawTextW").Call(dc, uintptr(unsafe.Pointer(&units[0])), uintptr(len(units)-1), uintptr(unsafe.Pointer(&r)), flags)
}
func line(dc uintptr, x1, y1, x2, y2 int32, color uint32) {
	pen, _, _ := gdi.NewProc("CreatePen").Call(0, 1, uintptr(color))
	old, _, _ := gdi.NewProc("SelectObject").Call(dc, pen)
	gdi.NewProc("MoveToEx").Call(dc, signed(x1), signed(y1), 0)
	gdi.NewProc("LineTo").Call(dc, signed(x2), signed(y2))
	gdi.NewProc("SelectObject").Call(dc, old)
	gdi.NewProc("DeleteObject").Call(pen)
}
func (a *launcherUI) showSettings(show bool) {
	a.settings = show
	for _, h := range a.settingControls {
		v := uintptr(0)
		if show {
			v = 5
		}
		user32.NewProc("ShowWindow").Call(h, v)
	}
	for _, h := range a.playControls {
		v := uintptr(5)
		if show {
			v = 0
		}
		user32.NewProc("ShowWindow").Call(h, v)
	}
	user32.NewProc("RedrawWindow").Call(a.hwnd, 0, 0, 0x85)
}
func paintTheme(hwnd uintptr) {
	type psType struct {
		dc                 uintptr
		erase              int32
		paint              rect
		restore, incUpdate int32
		reserved           [32]byte
	}
	var ps psType
	dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var client rect
	user32.NewProc("GetClientRect").Call(hwnd, uintptr(unsafe.Pointer(&client)))
	memory, _, _ := gdi.NewProc("CreateCompatibleDC").Call(dc)
	bitmap, _, _ := gdi.NewProc("CreateCompatibleBitmap").Call(dc, uintptr(client.right), uintptr(client.bottom))
	old, _, _ := gdi.NewProc("SelectObject").Call(memory, bitmap)
	drawScene(memory, client)
	gdi.NewProc("BitBlt").Call(dc, 0, 0, uintptr(client.right), uintptr(client.bottom), memory, 0, 0, 0x00CC0020)
	gdi.NewProc("SelectObject").Call(memory, old)
	gdi.NewProc("DeleteObject").Call(bitmap)
	gdi.NewProc("DeleteDC").Call(memory)
}
func drawScene(dc uintptr, client rect) {
	a := application
	fill(dc, client, colorBackground)
	fill(dc, rect{0, 0, 232, client.bottom}, 0x0A0E0B)
	line(dc, 232, 0, 232, client.bottom, colorLine)
	fill(dc, rect{24, 28, 28, 40}, colorGreen)
	text(dc, "IW5 / OPERATIONS", rect{38, 26, 214, 48}, colorMuted, theme.small, 0)
	text(dc, "SURVIVAL", rect{22, 66, 220, 110}, colorText, theme.subhead, 0)
	text(dc, "REIMAGINED", rect{24, 108, 220, 132}, colorGreen, theme.small, 0)
	line(dc, 24, 150, 208, 150, colorLine)
	text(dc, "COMPLETE EDITION", rect{24, 620, 220, 642}, colorMuted, theme.small, 0)
	text(dc, "v0.4.2 / WINDOWS", rect{24, 644, 220, 665}, colorMuted, theme.small, 0)
	text(dc, "IW5 / SURVIVAL REIMAGINED", rect{280, 28, 760, 50}, colorMuted, theme.small, 0)
	text(dc, "LAN / PRIVATE MATCH", rect{760, 28, 980, 50}, colorGreen, theme.small, 2)
	line(dc, 280, 66, 960, 66, colorLine)
	if a == nil {
		return
	}
	if a.settings {
		text(dc, "INSTELLINGEN", rect{278, 88, 960, 128}, colorText, theme.subhead, 0)
	} else {
		fill(dc, rect{260, 96, 980, 424}, colorPanel)
		for x := int32(620); x < 960; x += 24 {
			line(dc, x, 112, x, 404, 0x1B251C)
		}
		for y := int32(116); y < 408; y += 24 {
			line(dc, 604, y, 960, y, 0x1B251C)
		}
		for _, b := range []rect{{696, 180, 812, 260}, {804, 264, 926, 324}, {654, 294, 730, 354}} {
			line(dc, b.left, b.top, b.right, b.top, 0x41523A)
			line(dc, b.right, b.top, b.right, b.bottom, 0x41523A)
			line(dc, b.right, b.bottom, b.left, b.bottom, 0x41523A)
			line(dc, b.left, b.bottom, b.left, b.top, 0x41523A)
			line(dc, b.left, b.top, b.left+18, b.top-18, 0x41523A)
			line(dc, b.right, b.top, b.right+18, b.top-18, 0x41523A)
			line(dc, b.left+18, b.top-18, b.right+18, b.top-18, 0x41523A)
			line(dc, b.right+18, b.top-18, b.right+18, b.bottom-18, 0x41523A)
			line(dc, b.right+18, b.bottom-18, b.right, b.bottom, 0x41523A)
		}
		line(dc, 840, 118, 840, 394, colorLine)
		line(dc, 634, 246, 950, 246, colorLine)
		fill(dc, rect{836, 242, 844, 250}, colorGreen)
		text(dc, "SURVIVAL", rect{278, 126, 680, 210}, colorText, theme.display, 0)
		text(dc, "REIMAGINED", rect{282, 208, 700, 254}, colorGreen, theme.subhead, 0)
		text(dc, "HOUD STAND. BLIJF VECHTEN.", rect{284, 278, 700, 304}, colorMuted, theme.small, 0)
		text(dc, "Onbeperkte waves. Jouw loadout. Jouw strijd.", rect{284, 312, 700, 340}, colorText, theme.body, 0)
		mapName := strings.ToUpper(strings.TrimPrefix(windowText(a.mapName), "mp_"))
		text(dc, "STARTMAP / "+mapName, rect{284, 374, 700, 398}, colorGreen, theme.small, 0)
		text(dc, "Vervangen bestanden krijgen automatisch een backup.", rect{280, 532, 960, 556}, colorMuted, theme.body, 0)
	}
	line(dc, 280, 590, 960, 590, colorLine)
	fill(dc, rect{280, 590, 316, 592}, colorGreen)
	text(dc, "SYSTEEMSTATUS", rect{280, 566, 600, 588}, colorMuted, theme.small, 0)
	if a.busy {
		text(dc, "ACTIEF", rect{840, 566, 960, 588}, colorGreen, theme.small, 2)
	}
}
func drawThemedButton(pointer uintptr) {
	type drawItem struct {
		controlType, id, itemID, action, state uint32
		hwnd, dc                               uintptr
		bounds                                 rect
		data                                   uintptr
	}
	// LPARAM is numeric for many other messages. Keep the callback argument
	// an integer, and copy this OS-owned DRAWITEMSTRUCT into typed Go memory.
	var record drawItem
	syscall.NewLazyDLL("ntdll.dll").NewProc("RtlMoveMemory").Call(uintptr(unsafe.Pointer(&record)), pointer, unsafe.Sizeof(record))
	item := &record
	r := item.bounds
	primary := item.id == idStart || item.id == idSave
	selected := application != nil && ((item.id == idSettings && application.settings) || (item.id == idPlay && !application.settings))
	bg, fg := uint32(0x172019), colorText
	if primary {
		bg = colorGreen
		fg = colorBackground
	}
	if selected {
		bg = 0x223127
		fg = colorGreen
	}
	if item.state&1 != 0 {
		bg = 0x3D5947
		if primary {
			bg = 0x51AA90
		}
	}
	if item.state&4 != 0 {
		bg = 0x1B211D
		fg = 0x56645A
	}
	fill(item.dc, r, bg)
	if selected {
		fill(item.dc, rect{r.left, r.top, r.left + 3, r.bottom}, colorGreen)
	}
	f := theme.body
	if primary {
		f = theme.title
	}
	text(item.dc, windowText(item.hwnd), r, fg, f, 0x25)
	if item.state&0x10 != 0 {
		focus := rect{r.left + 3, r.top + 3, r.right - 3, r.bottom - 3}
		user32.NewProc("DrawFocusRect").Call(item.dc, uintptr(unsafe.Pointer(&focus)))
	}
}
