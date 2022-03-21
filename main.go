package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/ndbeals/winssh-pageant/internal/win"
)

var (
	sshPipe       = flag.String("sshpipe", `\\.\pipe\openssh-ssh-agent`, "Named pipe for Windows OpenSSH agent")
	noPageantPipe = flag.Bool("no-pageant-pipe", false, "Toggle pageant named pipe proxying")
)

var oldStdin, oldStdout, oldStderr *os.File

func main() {
	flag.Parse()

	err := win.FixConsoleIfNeeded()
	if err != nil {
		log.Fatalf("FixConsoleOutput: %v\n", err)
	}

	// Check if any application claiming to be a Pageant Window is already running
	if doesPagentWindowExist() {
		log.Println("This application is already running, exiting.")
		return
	}

	// Start a proxy/redirector for the pageant named pipes
	if !*noPageantPipe {
		go pipeProxy()
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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
}
