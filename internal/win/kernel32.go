package win

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// Library
	libkernel32 *windows.LazyDLL

	// Functions
	getLastError    *windows.LazyProc
	getModuleHandle *windows.LazyProc
	globalAlloc     *windows.LazyProc
	globalFree      *windows.LazyProc
)

func init() {
	// Library
	libkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// Functions
	getLastError = libkernel32.NewProc("GetLastError")
	getModuleHandle = libkernel32.NewProc("GetModuleHandleW")
	globalAlloc = libkernel32.NewProc("GlobalAlloc")
	globalFree = libkernel32.NewProc("GlobalFree")
}

func GetLastError() uint32 {
	ret, _, _ := syscall.Syscall(getLastError.Addr(), 0,
		0,
		0,
		0)

	return uint32(ret)
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
