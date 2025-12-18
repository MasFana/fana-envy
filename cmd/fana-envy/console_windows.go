//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func setupConsole() {
	handle := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err == nil {
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		windows.SetConsoleMode(handle, mode)
	}
}
