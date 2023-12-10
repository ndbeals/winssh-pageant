//go:build windows
// +build windows

package pageant

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/rs/zerolog/log"

	"encoding/binary"
	"encoding/hex"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"

	"github.com/ndbeals/winssh-pageant/internal/security"
	"github.com/ndbeals/winssh-pageant/internal/win"
	"github.com/ndbeals/winssh-pageant/openssh"
)

var defaultHandlerFunc = func(p *Pageant, result []byte) ([]byte, error) {
	return openssh.QueryAgent(p.SSHAgentPipe, result)
}

func (p *Pageant) Run() {
	// Check if any application claiming to be a Pageant Window is already running
	if doesPageantWindowExist() {
		log.Warn().Msg("This application is already running, exiting.")
		return
	}

	// Start a proxy/redirector for the pageant named pipes
	if p.pageantPipe {
		go p.pipeProxy()
		log.Info().Msg("Pageant pipe proxy started")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	log.Debug().Msg("Locked OS Thread")

	pageantWindow := p.createPageantWindow()
	if pageantWindow == 0 {
		// log.Println(fmt.Errorf("CreateWindowEx failed: %v", win.GetLastError()))
		log.Error().Stack().Err(win.GetLastError()).Msg("CreateWindowEx failed")
		return
	}

	hglobal := win.GlobalAlloc(0, unsafe.Sizeof(win.MSG{}))
	//nolint:gosec
	msg := (*win.MSG)(unsafe.Pointer(hglobal))
	log.Debug().Msg("Allocated global memory for message data, Starting message loop")

	// main message loop
	for win.GetMessage(msg, pageantWindow, 0, 0) > 0 {
		win.TranslateMessage(msg)
		win.DispatchMessage(msg)
	}

	log.Debug().Msg("Message loop exited, freeing global memory")
	// Explicitly release the global memory handle
	win.GlobalFree(hglobal)
}

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
	if err != nil {
		log.Error().Stack().Err(err).Msg("OpenFileMapping syscall failed")
	}
	return windows.Handle(mapPtr), err
}

func doesPageantWindowExist() bool {
	return win.FindWindow(wndClassNamePtr, nil) != 0
}

func (p *Pageant) registerPageantWindow(hInstance win.HINSTANCE) (atom win.ATOM) {
	var wc win.WNDCLASSEX
	wc.Style = 0

	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(p.wndProc)
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

func (p *Pageant) createPageantWindow() win.HWND {
	inst := win.GetModuleHandle(nil)
	atom := p.registerPageantWindow(inst)
	if atom == 0 {
		// log.Println(fmt.Errorf("RegisterClass failed: %d", win.GetLastError()))
		log.Error().Stack().Err(win.GetLastError()).Msg("RegisterClass failed")
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

func (p *Pageant) wndProc(hWnd win.HWND, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case win.WM_COPYDATA:
		{
			copyData := (*copyDataStruct)(unsafe.Pointer(lParam))
			if copyData.dwData != agentCopyDataID {
				return 0
			}
			log.Debug().Msg("Received WM_COPYDATA message")

			fileMap, err := openFileMap(FILE_MAP_ALL_ACCESS, 0, copyData.lpData)
			if err != nil {
				// log.Println(err)
				log.Error().Stack().Err(err).Msg("OpenFileMap failed")
				return 0
			}
			defer func() {
				err := windows.CloseHandle(fileMap)
				if err != nil {
					log.Error().Stack().Err(err).Msg("CloseHandle failed")
				}
			}()

			// check security
			ourself, err := security.GetUserSID()
			if err != nil {
				// log.Println(err)
				log.Error().Stack().Err(err).Msg("GetUserSID failed")
				return 0
			}
			ourself2, err := security.GetDefaultSID()
			if err != nil {
				// log.Println(err)
				log.Error().Stack().Err(err).Msg("GetDefaultSID failed")
				return 0
			}
			mapOwner, err := security.GetHandleSID(fileMap)
			if err != nil {
				// log.Println(err)
				log.Error().Stack().Err(err).Msg("GetHandleSID failed")
				return 0
			}
			if !windows.EqualSid(mapOwner, ourself) && !windows.EqualSid(mapOwner, ourself2) {
				return 0
			}
			log.Debug().Msg("Passed security checks")

			// Passed security checks, copy data
			sharedMemory, err := windows.MapViewOfFile(fileMap, FILE_MAP_WRITE, 0, 0, 0)
			if err != nil {
				// log.Println(err)
				log.Error().Stack().Err(err).Msg("MapViewOfFile failed")
				return 0
			}
			defer func() {
				err := windows.UnmapViewOfFile(sharedMemory)
				if err != nil {
					log.Error().Stack().Err(err).Msg("UnmapViewOfFile failed")
				}
			}()

			sharedMemoryArray := (*[openssh.AgentMaxMessageLength]byte)(unsafe.Pointer(sharedMemory))

			size := binary.BigEndian.Uint32(sharedMemoryArray[:4]) + 4 // +4 for the size uint itself
			if size > openssh.AgentMaxMessageLength {
				return 0
			}
			log.Debug().Msgf("Received pageant request message of size: %d", size)

			// Query the windows OpenSSH agent via the windows named pipe
			result, err := p.PageantRequestHandler(p, sharedMemoryArray[:size])
			if err != nil {
				log.Printf("Error in PageantRequestHandler: %+v\n", err)
				return 0
			}
			log.Debug().Msgf("Sent request to openssh handler, result size: %d", len(result))
			elems := copy(sharedMemoryArray[:], result)
			log.Debug().Msgf("Copied %d elements to shared memory", elems)

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

func (p *Pageant) pipeProxy() {
	currentUser, err := user.Current()
	if err != nil {
		// log.Println(err)
		log.Error().Stack().Err(err).Msg("user.Current failed")
	}

	namePart := strings.Split(currentUser.Username, `\`)[1]
	pipeName := fmt.Sprintf(agentPipeName, namePart, capiObfuscateString(wndClassName))
	listener, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		// log.Println(err)
		log.Error().Stack().Err(err).Msg("ListenPipe failed")
	} else {
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				// log.Println(err)
				log.Error().Stack().Err(err).Msg("Pipe proxy accept failed")
				return
			}
			go p.pipeListen(conn)
		}
	}
}

func (p *Pageant) pipeListen(pageantConn net.Conn) {
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

		result, err := p.PageantRequestHandler(p, append(lenBuf, readBuf...))
		if err != nil {
			log.Printf("Pipe: Error in PageantRequestHandler: %+v\n", err)
			return
		}

		_, err = pageantConn.Write(result)
		if err != nil {
			return
		}
	}
}
