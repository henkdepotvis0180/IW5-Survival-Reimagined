package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

func lockLaunch() (func(), error) {
	h, _, err := kernel32.NewProc("CreateMutexW").Call(0, 0, uintptr(unsafe.Pointer(wide(`Local\IW5SurvivalReimaginedLaunch`))))
	if h == 0 {
		return nil, fmt.Errorf("Startvergrendeling mislukt: %v", err)
	}
	if err == syscall.Errno(183) {
		syscall.CloseHandle(syscall.Handle(h))
		return nil, fmt.Errorf("Een andere Survival-launcher start de game al of heeft deze nog geopend")
	}
	// Holding the named object open acts as a process-wide launch lease; no
	// thread ownership is used because Go goroutines may switch OS threads.
	return func() { syscall.CloseHandle(syscall.Handle(h)) }, nil
}
func requireGameClosed() error {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(snapshot)
	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	err = syscall.Process32First(snapshot, &entry)
	for err == nil {
		name := strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:]))
		if name == "plutonium-bootstrapper-win32.exe" || name == "iw5mp.exe" || name == "iw5mp_server.exe" {
			return fmt.Errorf("Sluit de actieve Plutonium/MW3-game voordat instellingen of modbestanden worden bijgewerkt (PID %d)", entry.ProcessID)
		}
		err = syscall.Process32Next(snapshot, &entry)
	}
	if err != syscall.ERROR_NO_MORE_FILES {
		return err
	}
	return nil
}
