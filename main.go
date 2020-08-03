package main

import (
	"crypto/sha256"
	"flag"
	"fmt"

	"github.com/lxn/win"
)

var (
	sshPipe = flag.String("sshpipe", `\\.\pipe\openssh-ssh-agent`, "Named pipe for Windows OpenSSH agent")
)

// var testPipeName = `\\.\pipe\pageant.Nate.de663e60ca4268cef1d1f931f1d8ffe2d0df1de50aa09d8cc29facc5991d3638`
var testPipeName = `\\.\pipe\pageant.Nate`

func main() {
	flag.Parse()

	fmt.Println(fmt.Sprintf(`terst\%s.%s`, "Nate", "sha"))
	fmt.Println("open serv")
	go pipeProxy()
	fmt.Println("postopen serv")

	fmt.Println("wait rec:")
	// fmt.Println(<-ch)

	sum := sha256.Sum256([]byte(`Pageant`))
	fmt.Printf("%x", sum)

	pageantWindow := createPageantWindow()
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
