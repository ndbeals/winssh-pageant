package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"github.com/ndbeals/winssh-pageant/internal/win"
)

var (
	sshPipe       = flag.String("sshpipe", `\\.\pipe\openssh-ssh-agent`, "Named pipe for Windows OpenSSH agent")
	noPageantPipe = flag.Bool("no-pageant-pipe", false, "Toggle pageant named pipe proxying")
)

func main() {
	flag.Parse()

	// Start a proxy/redirector for the pageant named pipes
	if !*noPageantPipe {
		go pipeProxy()
	}

	runtime.LockOSThread()

	pageantWindow := createPageantWindow()
	if pageantWindow == 0 {
		log.Println(fmt.Errorf("CreateWindowEx failed: %v", win.GetLastError()))
		return
	}

	hglobal := win.GlobalAlloc(0, unsafe.Sizeof(win.MSG{}))
	msg := (*win.MSG)(unsafe.Pointer(hglobal))

	// main message loop
	for win.GetMessage(msg, 0, 0, 0) > 0 {
		win.TranslateMessage(msg)
		win.DispatchMessage(msg)
	}

	// Explicitly release the global memory handle
	win.GlobalFree(hglobal)
	runtime.UnlockOSThread()
}
