package main

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/lxn/win"
)

var (
	sshPipe = flag.String("sshpipe", `\\.\pipe\openssh-ssh-agent`, "Named pipe for Windows OpenSSH agent")
)

func main() {
	flag.Parse()

	inst := win.GetModuleHandle(nil)
	atom := registerPageantWindow(inst)
	if atom == 0 {
		fmt.Println(fmt.Errorf("RegisterClass failed: %d", win.GetLastError()))
		return
	}

	// CreateWindowEx
	pageantWindow := win.CreateWindowEx(win.WS_EX_APPWINDOW,
		syscall.StringToUTF16Ptr(wndClassName),
		syscall.StringToUTF16Ptr(wndClassName),
		0,
		0, 0,
		0, 0,
		0,
		0,
		inst,
		nil)
	if pageantWindow == 0 {
		fmt.Println(fmt.Errorf("CreateWindowEx failed: %v", win.GetLastError()))
		return
	}

	// main message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}
