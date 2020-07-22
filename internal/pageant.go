package utils

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

func WinMain(Inst win.HINSTANCE) int32 {
	// RegisterClass
	atom := registerWindowClass(Inst)
	if atom == 0 {
		fmt.Errorf("RegisterClass failed: %d", win.GetLastError())
		return 0
	}

	// CreateWindowEx
	wnd := win.CreateWindowEx(win.WS_EX_APPWINDOW,
		syscall.StringToUTF16Ptr(wndClassName),
		syscall.StringToUTF16Ptr(wndClassName),
		0,
		0, 0,
		0, 0,
		0,
		0,
		Inst,
		nil)
	if wnd == 0 {
		fmt.Errorf("CreateWindowEx failed: %v", win.GetLastError())
		return 0
	}

	// main message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	return int32(msg.WParam)
}

func WndProc(hWnd win.HWND, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case win.WM_COPYDATA:
		{
			copyData := (*copyDataStruct)(unsafe.Pointer(lParam))

			fileMap, err := openFileMap(FILE_MAP_ALL_ACCESS, 0, copyData.lpData)
			defer windows.CloseHandle(fileMap)

			// check security
			ourself, err := GetUserSID()
			if err != nil {
				return 0
			}
			ourself2, err := GetDefaultSID()
			if err != nil {
				return 0
			}
			mapOwner, err := GetHandleSID(fileMap)
			if err != nil {
				return 0
			}
			if !windows.EqualSid(mapOwner, ourself) && !windows.EqualSid(mapOwner, ourself2) {
				return 0
			}

			sharedMemory, err := windows.MapViewOfFile(fileMap, 2, 0, 0, 0)
			if err != nil {
				return 0
			}
			defer windows.UnmapViewOfFile(sharedMemory)

			sharedMemoryArray := (*[agentMaxMessageLength]byte)(unsafe.Pointer(sharedMemory))

			size := binary.BigEndian.Uint32(sharedMemoryArray[:4])
			size += 4
			if size > agentMaxMessageLength {
				return 0
			}

			result, err := queryAgent("\\\\.\\pipe\\openssh-ssh-agent", sharedMemoryArray[:size])
			copy(sharedMemoryArray[:], result)

			// success
			return 1
		}
	}

	return win.DefWindowProc(hWnd, message, wParam, lParam)
}
