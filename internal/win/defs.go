package win

const (
	//revive:disable:var-naming,exported
	IDI_APPLICATION       = 32512
	IDC_IBEAM             = 32513
	BLACK_BRUSH           = 4
	WM_COPYDATA           = 74
	WM_DESTROY            = 2
	WM_CLOSE              = 16
	WM_QUERYENDSESSION    = 17
	WM_QUIT               = 18
	WS_EX_TOOLWINDOW      = 0x00000080
	WS_EX_APPWINDOW       = 0x00040000
	ATTACH_PARENT_PROCESS = uintptr(^uint32(0)) // (DWORD)-1
	ERROR_INVALID_HANDLE  = 6
)

type (
	BOOL      int32
	HRESULT   int32
	ATOM      uint16
	HANDLE    uintptr
	HGLOBAL   HANDLE
	HINSTANCE HANDLE
	HACCEL    HANDLE
	HCURSOR   HANDLE
	HDWP      HANDLE
	HICON     HANDLE
	HMENU     HANDLE
	HMONITOR  HANDLE
	HRAWINPUT HANDLE
	HWND      HANDLE
	HBRUSH    HGDIOBJ
	HGDIOBJ   HANDLE
)

type MSG struct {
	HWnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

type POINT struct {
	X, Y int32
}
