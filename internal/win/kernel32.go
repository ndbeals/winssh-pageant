package win

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// Library
	libkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	// Functions
	getLastError    = libkernel32.NewProc("GetLastError")
	getModuleHandle = libkernel32.NewProc("GetModuleHandleW")
	globalAlloc     = libkernel32.NewProc("GlobalAlloc")
	globalFree      = libkernel32.NewProc("GlobalFree")
	attachConsole   = libkernel32.NewProc("AttachConsole")
)

// func GetLastError() uint32 {
// 	ret, _, _ := syscall.Syscall(getLastError.Addr(), 0,
// 		0,
// 		0,
// 		0)

// 	return uint32(ret)
// }

func GetLastError() (lasterr error) {
	r0, _, _ := syscall.SyscallN(getLastError.Addr())
	if r0 != 0 {
		lasterr = syscall.Errno(r0)
	}
	return lasterr
}

func GetModuleHandle(lpModuleName *uint16) HINSTANCE {
	ret, _, _ := syscall.Syscall(getModuleHandle.Addr(), 1,
		uintptr(unsafe.Pointer(lpModuleName)),
		0,
		0)

	return HINSTANCE(ret)
}

func GlobalAlloc(uFlags uint32, dwBytes uintptr) HGLOBAL {
	ret, _, _ := syscall.Syscall(globalAlloc.Addr(), 2,
		uintptr(uFlags),
		dwBytes,
		0)

	return HGLOBAL(ret)
}

func GlobalFree(hMem HGLOBAL) HGLOBAL {
	ret, _, _ := syscall.Syscall(globalFree.Addr(), 1,
		uintptr(hMem),
		0,
		0)

	return HGLOBAL(ret)
}
