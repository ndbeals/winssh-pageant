package win

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// Library
	libuser32 = windows.NewLazySystemDLL("user32.dll")
	// Functions
	loadCursor       = libuser32.NewProc("LoadCursorW")
	loadIcon         = libuser32.NewProc("LoadIconW")
	getSysColorBrush = libuser32.NewProc("GetSysColorBrush")
	registerClassEx  = libuser32.NewProc("RegisterClassExW")
	createWindowEx   = libuser32.NewProc("CreateWindowExW")
	defWindowProc    = libuser32.NewProc("DefWindowProcW")
	getMessage       = libuser32.NewProc("GetMessageW")
	translateMessage = libuser32.NewProc("TranslateMessage")
	dispatchMessage  = libuser32.NewProc("DispatchMessageW")
	postQuitMessage  = libuser32.NewProc("PostQuitMessage")
)

func MAKEINTRESOURCE(id uintptr) *uint16 {
	return (*uint16)(unsafe.Pointer(id))
}

func LoadCursor(hInstance HINSTANCE, lpCursorName *uint16) HCURSOR {
	ret, _, _ := syscall.Syscall(loadCursor.Addr(), 2,
		uintptr(hInstance),
		uintptr(unsafe.Pointer(lpCursorName)),
		0)

	return HCURSOR(ret)
}

func LoadIcon(hInstance HINSTANCE, lpIconName *uint16) HICON {
	ret, _, _ := syscall.Syscall(loadIcon.Addr(), 2,
		uintptr(hInstance),
		uintptr(unsafe.Pointer(lpIconName)),
		0)

	return HICON(ret)
}

func GetSysColorBrush(nIndex int) HBRUSH {
	ret, _, _ := syscall.Syscall(getSysColorBrush.Addr(), 1,
		uintptr(nIndex),
		0,
		0)

	return HBRUSH(ret)
}

func RegisterClassEx(windowClass *WNDCLASSEX) ATOM {
	ret, _, _ := syscall.Syscall(registerClassEx.Addr(), 1,
		uintptr(unsafe.Pointer(windowClass)),
		0,
		0)

	return ATOM(ret)
}

func CreateWindowEx(dwExStyle uint32, lpClassName, lpWindowName *uint16, dwStyle uint32, x, y, nWidth, nHeight int32, hWndParent HWND, hMenu HMENU, hInstance HINSTANCE, lpParam unsafe.Pointer) HWND {
	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
		uintptr(dwExStyle),
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)),
		uintptr(dwStyle),
		uintptr(x),
		uintptr(y),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(hWndParent),
		uintptr(hMenu),
		uintptr(hInstance),
		uintptr(lpParam))

	return HWND(ret)
}

func DefWindowProc(hWnd HWND, Msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := syscall.Syscall6(defWindowProc.Addr(), 4,
		uintptr(hWnd),
		uintptr(Msg),
		wParam,
		lParam,
		0,
		0)

	return ret
}

func GetMessage(msg *MSG, hWnd HWND, msgFilterMin, msgFilterMax uint32) BOOL {
	ret, _, _ := syscall.Syscall6(getMessage.Addr(), 4,
		uintptr(unsafe.Pointer(msg)),
		uintptr(hWnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
		0,
		0)

	return BOOL(ret)
}

func TranslateMessage(msg *MSG) bool {
	ret, _, _ := syscall.Syscall(translateMessage.Addr(), 1,
		uintptr(unsafe.Pointer(msg)),
		0,
		0)

	return ret != 0
}

func DispatchMessage(msg *MSG) uintptr {
	ret, _, _ := syscall.Syscall(dispatchMessage.Addr(), 1,
		uintptr(unsafe.Pointer(msg)),
		0,
		0)

	return ret
}

func PostQuitMessage(exitCode int32) {
	syscall.Syscall(postQuitMessage.Addr(), 1,
		uintptr(exitCode),
		0,
		0)
}
