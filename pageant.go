package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os/user"
	"strings"
	"syscall"
	"unsafe"

	"encoding/binary"

	"github.com/Microsoft/go-winio"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"github.com/ndbeals/winssh-pageant/internal/security"
	"github.com/ndbeals/winssh-pageant/internal/sshagent"
)

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procOpenFileMappingA = modkernel32.NewProc("OpenFileMappingA")
)

const (
	// windows consts
	FILE_MAP_ALL_ACCESS = 0xf001f

	// Pageant consts
	agentPipeName   = `\\.\pipe\pageant.%s.%x`
	agentCopyDataID = 0x804e50ba
	wndClassName    = "Pageant"
)

// copyDataStruct is used to pass data in the WM_COPYDATA message.
// We directly pass a pointer to our copyDataStruct type, be careful that it matches the Windows type exactly
type copyDataStruct struct {
	dwData uintptr
	cbData uint32
	lpData uintptr
}

func registerPageantWindow(hInstance win.HINSTANCE) (atom win.ATOM) {
	var wc win.WNDCLASSEX
	wc.Style = 0

	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.CbClsExtra = 0
	wc.CbWndExtra = 0
	wc.HInstance = hInstance
	wc.HIcon = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))
	wc.HCursor = win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_IBEAM))
	wc.HbrBackground = win.GetSysColorBrush(win.BLACK_BRUSH)
	wc.LpszMenuName = nil
	wc.LpszClassName = syscall.StringToUTF16Ptr(wndClassName)
	wc.HIconSm = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))

	return win.RegisterClassEx(&wc)
}

func createPageantWindow() win.HWND {
	inst := win.GetModuleHandle(nil)
	atom := registerPageantWindow(inst)
	if atom == 0 {
		fmt.Println(fmt.Errorf("RegisterClass failed: %d", win.GetLastError()))
		return 0
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

	return pageantWindow
}

func openFileMap(dwDesiredAccess uint32, bInheritHandle uint32, mapNamePtr uintptr) (windows.Handle, error) {
	mapPtr, _, err := procOpenFileMappingA.Call(uintptr(dwDesiredAccess), uintptr(bInheritHandle), mapNamePtr)

	if err != nil && err.Error() == "The operation completed successfully." {
		err = nil
	}

	return windows.Handle(mapPtr), err
}

func wndProc(hWnd win.HWND, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case win.WM_COPYDATA:
		{
			copyData := (*copyDataStruct)(unsafe.Pointer(lParam))

			fileMap, err := openFileMap(FILE_MAP_ALL_ACCESS, 0, copyData.lpData)
			defer windows.CloseHandle(fileMap)

			// check security
			ourself, err := security.GetUserSID()
			if err != nil {
				return 0
			}
			ourself2, err := security.GetDefaultSID()
			if err != nil {
				return 0
			}
			mapOwner, err := security.GetHandleSID(fileMap)
			if err != nil {
				return 0
			}
			if !windows.EqualSid(mapOwner, ourself) && !windows.EqualSid(mapOwner, ourself2) {
				return 0
			}

			// Passed security checks, copy data
			sharedMemory, err := windows.MapViewOfFile(fileMap, 2, 0, 0, 0)
			if err != nil {
				return 0
			}
			defer windows.UnmapViewOfFile(sharedMemory)

			sharedMemoryArray := (*[sshagent.AgentMaxMessageLength]byte)(unsafe.Pointer(sharedMemory))

			size := binary.BigEndian.Uint32(sharedMemoryArray[:4]) + 4
			// size += 4
			if size > sshagent.AgentMaxMessageLength {
				return 0
			}

			// result, err := sshagent.QueryAgent(*sshPipe, sharedMemoryArray[:size], sshagent.AgentMaxMessageLength)
			result, err := sshagent.QueryAgent(*sshPipe, sharedMemoryArray[:size])
			copy(sharedMemoryArray[:], result)
			// success
			return 1
		}
	}

	return win.DefWindowProc(hWnd, message, wParam, lParam)
}

func pipeProxy() {
	currentUser, err := user.Current()
	pipeName := fmt.Sprintf(agentPipeName, strings.Split(currentUser.Username, `\`)[1], sha256.Sum256([]byte(wndClassName)))
	listener, err := winio.ListenPipe(pipeName, nil)

	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go pipeListen(conn)
	}
}

func pipeListen(pageantConn net.Conn) {
	defer pageantConn.Close()
	reader := bufio.NewReader(pageantConn)

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lenBuf)
		if err != nil {
			// if *verbose {
			// 	log.Printf("io.ReadFull error '%s'", err)
			// }
			return
		}

		bufferLen := binary.BigEndian.Uint32(lenBuf)
		readBuf := make([]byte, bufferLen)
		_, err = io.ReadFull(reader, readBuf)
		if err != nil {
			// if *verbose {
			// 	log.Printf("io.ReadFull error '%s'", err)
			// }
			return
		}

		result, err := sshagent.QueryAgent(*sshPipe, append(lenBuf, readBuf...))
		if err != nil {
			// if *verbose {
			// 	log.Printf("Pageant query error '%s'", err)
			// }
			// result = failureMessage[:]
		}

		_, err = pageantConn.Write(result)
		if err != nil {
			// if *verbose {
			// 	log.Printf("net.Conn.Write error '%s'", err)
			// }
			return
		}
	}
}
