//go:build windows
// +build windows

package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	cfUnicodetext = 13
	gmemMoveable  = 0x0002
)

var (
	user32               = syscall.MustLoadDLL("user32")
	openClipboardProc    = user32.MustFindProc("OpenClipboard")
	closeClipboardProc   = user32.MustFindProc("CloseClipboard")
	emptyClipboardProc   = user32.MustFindProc("EmptyClipboard")
	getClipboardProc     = user32.MustFindProc("GetClipboardData")
	setClipboardDataProc = user32.MustFindProc("SetClipboardData")

	kernel32     = syscall.NewLazyDLL("kernel32")
	globalAlloc  = kernel32.NewProc("GlobalAlloc")
	globalFree   = kernel32.NewProc("GlobalFree")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
	lstrcpy      = kernel32.NewProc("lstrcpyW")
)

func openClipboard() error {
	started := time.Now()
	limit := started.Add(time.Second)
	var r uintptr
	var err error
	for time.Now().Before(limit) {
		r, _, err = openClipboardProc.Call(0)
		if r != 0 {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
	return err
}

func writeToClipboard(text string) error {
	var err error
	if err = openClipboard(); err != nil {
		return err
	}
	defer closeClipboardProc.Call()
	var r uintptr
	if r, _, err = emptyClipboardProc.Call(0); r == 0 {
		fmt.Println("Clipboard konnte nicht geloescht werden!")
		return err
	}

	var data []uint16
	if data, err = syscall.UTF16FromString(text); err != nil {
		return err
	}
	h, _, err := globalAlloc.Call(gmemMoveable, uintptr(len(data)*int(unsafe.Sizeof(data[0]))))
	if h == 0 {
		return err
	}
	defer func() {
		if h != 0 {
			globalFree.Call(h)
		}
	}()

	l, _, err := globalLock.Call(h)
	if l == 0 {
		return err
	}

	r, _, err = lstrcpy.Call(l, uintptr(unsafe.Pointer(&data[0])))
	if r == 0 {
		return err
	}

	r, _, err = globalUnlock.Call(h)
	if r == 0 {
		if err.(syscall.Errno) != 0 {
			return err
		}
	}

	r, _, err = setClipboardDataProc.Call(cfUnicodetext, h)
	if r == 0 {
		return err
	}
	h = 0 // suppress deferred cleanup
	return nil
}
