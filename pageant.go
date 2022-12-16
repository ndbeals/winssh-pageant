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
	"encoding/hex"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"

	"github.com/ndbeals/winssh-pageant/internal/security"
	"github.com/ndbeals/winssh-pageant/internal/sshagent"
	"github.com/ndbeals/winssh-pageant/internal/win"
)

const (
	// windows consts
	//revive:disable:var-naming,exported
	CRYPTPROTECTMEMORY_BLOCK_SIZE    = 16
	CRYPTPROTECTMEMORY_CROSS_PROCESS = 1
	FILE_MAP_ALL_ACCESS              = 0xf001f
	FILE_MAP_WRITE                   = 0x2

	// Pageant consts
	agentPipeName   = `\\.\pipe\pageant.%s.%s`
	agentCopyDataID = 0x804e50ba
	wndClassName    = "Pageant"
)

var (
	crypt32                = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectMemory = crypt32.NewProc("CryptProtectMemory")

	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procOpenFileMappingA = modkernel32.NewProc("OpenFileMappingA")
	wndClassNamePtr, _   = syscall.UTF16PtrFromString(wndClassName)
)

// copyDataStruct is used to pass data in the WM_COPYDATA message.
// We directly pass a pointer to our copyDataStruct type, be careful that it matches the Windows type exactly
type copyDataStruct struct {
	dwData uintptr
	cbData uint32
	lpData uintptr
}

func openFileMap(dwDesiredAccess, bInheritHandle uint32, mapNamePtr uintptr) (windows.Handle, error) {
	mapPtr, _, err := procOpenFileMappingA.Call(uintptr(dwDesiredAccess), uintptr(bInheritHandle), mapNamePtr)

	//Properly compare syscall.Errno to number, instead of naive (i18n-unaware) string comparison
	if err.(syscall.Errno) == windows.ERROR_SUCCESS {
		err = nil
	}
	return windows.Handle(mapPtr), err
}

func doesPagentWindowExist() bool {
	return win.FindWindow(wndClassNamePtr, nil) != 0
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
	wc.LpszClassName = wndClassNamePtr
	wc.HIconSm = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))

	return win.RegisterClassEx(&wc)
}

func createPageantWindow() win.HWND {
	inst := win.GetModuleHandle(nil)
	atom := registerPageantWindow(inst)
	if atom == 0 {
		log.Println(fmt.Errorf("RegisterClass failed: %d", win.GetLastError()))
		return 0
	}

	// CreateWindowEx
	pageantWindow := win.CreateWindowEx(
		win.WS_EX_APPWINDOW,
		wndClassNamePtr,
		wndClassNamePtr,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		inst,
		nil,
	)

	return pageantWindow
}

func wndProc(hWnd win.HWND, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case win.WM_COPYDATA:
		{
			copyData := (*copyDataStruct)(unsafe.Pointer(lParam))
			if copyData.dwData != agentCopyDataID {
				return 0
			}

			fileMap, err := openFileMap(FILE_MAP_ALL_ACCESS, 0, copyData.lpData)
			if err != nil {
				log.Println(err)
				return 0
			}
			defer windows.CloseHandle(fileMap)

			// check security
			ourself, err := security.GetUserSID()
			if err != nil {
				log.Println(err)
				return 0
			}
			ourself2, err := security.GetDefaultSID()
			if err != nil {
				log.Println(err)
				return 0
			}
			mapOwner, err := security.GetHandleSID(fileMap)
			if err != nil {
				log.Println(err)
				return 0
			}
			if !windows.EqualSid(mapOwner, ourself) && !windows.EqualSid(mapOwner, ourself2) {
				return 0
			}

			// Passed security checks, copy data
			sharedMemory, err := windows.MapViewOfFile(fileMap, FILE_MAP_WRITE, 0, 0, 0)
			if err != nil {
				log.Println(err)
				return 0
			}
			defer windows.UnmapViewOfFile(sharedMemory)

			sharedMemoryArray := (*[sshagent.AgentMaxMessageLength]byte)(unsafe.Pointer(sharedMemory))

			size := binary.BigEndian.Uint32(sharedMemoryArray[:4]) + 4 // +4 for the size uint itself
			if size > sshagent.AgentMaxMessageLength {
				return 0
			}

			// Query the windows OpenSSH agent via the windows named pipe
			result, err := sshagent.QueryAgent(*sshPipe, sharedMemoryArray[:size])
			if err != nil {
				log.Println(err)
				return 0
			}
			copy(sharedMemoryArray[:], result)

			// success, explicitly Clean up some resources (better to be certain it get's GC'd)
			ourself = nil
			ourself2 = nil
			mapOwner = nil
			sharedMemoryArray = nil
			result = nil

			return 1
		}
	case win.WM_DESTROY, win.WM_CLOSE, win.WM_QUIT, win.WM_QUERYENDSESSION:
		{ // Handle system shutdowns and process sigterms etc
			win.PostQuitMessage(0)
			return 0
		}
	}

	return win.DefWindowProc(hWnd, message, wParam, lParam)
}

func capiObfuscateString(realname string) string {
	cryptlen := len(realname) + 1
	cryptlen += CRYPTPROTECTMEMORY_BLOCK_SIZE - 1
	cryptlen /= CRYPTPROTECTMEMORY_BLOCK_SIZE
	cryptlen *= CRYPTPROTECTMEMORY_BLOCK_SIZE

	cryptdata := make([]byte, cryptlen)
	copy(cryptdata, realname)

	pDataIn := uintptr(unsafe.Pointer(&cryptdata[0]))
	cbDataIn := uintptr(cryptlen)
	dwFlags := uintptr(CRYPTPROTECTMEMORY_CROSS_PROCESS)

	//revive:disable:unhandled-error  - pageant ignores errors
	procCryptProtectMemory.Call(pDataIn, cbDataIn, dwFlags)

	hash := sha256.Sum256(cryptdata)
	return hex.EncodeToString(hash[:])
}

func pipeProxy() {
	currentUser, err := user.Current()
	if err != nil {
		log.Println(err)
	}

	namePart := strings.Split(currentUser.Username, `\`)[1]
	pipeName := fmt.Sprintf(agentPipeName, namePart, capiObfuscateString(wndClassName))
	listener, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		log.Println(err)
	} else {
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
}

func pipeListen(pageantConn net.Conn) {
	defer pageantConn.Close()
	reader := bufio.NewReader(pageantConn)

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lenBuf)
		if err != nil {
			return
		}

		bufferLen := binary.BigEndian.Uint32(lenBuf)
		readBuf := make([]byte, bufferLen)
		_, err = io.ReadFull(reader, readBuf)
		if err != nil {
			return
		}

		result, err := sshagent.QueryAgent(*sshPipe, append(lenBuf, readBuf...))
		if err != nil {
			return
		}

		_, err = pageantConn.Write(result)
		if err != nil {
			return
		}
	}
}
