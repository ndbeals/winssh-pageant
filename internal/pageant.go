package utils

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

func WinMain(Inst win.HINSTANCE) int32 {
	// RegisterClass
	atom := MyRegisterClass(Inst)
	if atom == 0 {
		fmt.Println("RegisterClass failed:", win.GetLastError())
		return 0
	}
	fmt.Println("RegisterClass ok", atom)

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
		fmt.Println("CreateWindowEx failed:", win.GetLastError())
		return 0
	}
	fmt.Println("CreateWindowEx done", wnd)
	// win.ShowWindow(wnd, win.SW_SHOW)
	// win.UpdateWindow(wnd)

	// main message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	return int32(msg.WParam)
}

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procOpenFileMappingW = modkernel32.NewProc("OpenFileMappingW")
	procOpenFileMappingA = modkernel32.NewProc("OpenFileMappingA")
)

func WndProc(hWnd win.HWND, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case win.WM_COPYDATA:
		{
			// cds := unsafe.Pointer(lParam)
			cds := (*copyDataStruct)(unsafe.Pointer(lParam))

			fmt.Printf("wndProc COPYDATA: %+v %+v %+v \n", hWnd, wParam, lParam)

			// nameLength := cds.cbData //(*reflect.StringHeader)(unsafe.Pointer(uintptr(cds.cbData)))
			// name := (*reflect.StringHeader)(unsafe.Pointer(uintptr(cds.lpData)))
			// name := (*string)(unsafe.Pointer(cds.lpData))
			// name := (*char)(unsafe.Pointer(cds.lpData))
			// name := make([]uint16,nameLength)
			// var data *byte = (*byte)(unsafe.Pointer(cds.lpData))
			// mapName := fmt.Sprintf( utf16PtrToString(data,nameLength) )
			// mapName := utf16PtrToString(data, nameLength)
			// sa := makeInheritSaWithSid()

			// fileMap, err := windows.CreateFileMapping(invalidHandleValue, sa, pageReadWrite, 0, agentMaxMessageLength, syscall.StringToUTF16Ptr(mapName))
			// fileMap,err := syscall.Open(mapName,0,0)
			// fileMap,_,err := syscall.Syscall(procOpenFileMappingW.Addr(),3,4,1,uintptr(*syscall.StringToUTF16Ptr(mapName)))
			// fileMapp, _, err := procOpenFileMappingW.Call(uintptr(983071), uintptr(0), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(mapName))))
			fileMapp, _, err := procOpenFileMappingA.Call(uintptr(FILE_MAP_ALL_ACCESS), uintptr(0), cds.lpData)
			// fmt.Printf("exists: %d \n", win.GetLastError())
			// if err != nil {
			// 	fmt.Errorf("shit. %d \n",win.GetLastError())
			// }
			fileMap := windows.Handle(fileMapp)
			// fileMap := (windows.Handle)(unsafe.Pointer(fileMapp))
			defer func() {
				windows.CloseHandle(fileMap)
				// 	// queryPageantMutex.Unlock()
			}()
			if err.Error() != "The operation completed successfully." {
				fmt.Println(err.Error())
				fmt.Errorf("Error on open, \n")
				return 0
			}

			// // const unt
			sharedMemory, err := windows.MapViewOfFile(fileMap, 2, 0, 0, 0)
			if err != nil {
				return 0
			}
			defer windows.UnmapViewOfFile(sharedMemory)

			sharedMemoryArray := (*[agentMaxMessageLength]byte)(unsafe.Pointer(sharedMemory))
			// // var buf []byte //= make([]byte)
			// // copy(sharedMemoryArray[:], buf)
			// // copy(buf, sharedMemoryArray[:])
			// // lenBuf := len(buf)
			// lenBuf := sharedMemoryArray[0:4]
			leng := binary.BigEndian.Uint32(sharedMemoryArray[:4])
			leng += 4
			bb := sharedMemoryArray[200:210]
			// leng :=0
			// bb:=0
			if leng > agentMaxMessageLength {
				return 0
			}
			fmt.Printf("cds: %+v %+v %+v %+v %+v %+v", fileMap, leng, bb)
			// leng = 4096

			// dataa := make([]byte, leng)
			dataa := sharedMemoryArray[:leng]
			fmt.Println("before quiery agent")
			result, err := queryAgent("\\\\.\\pipe\\openssh-ssh-agent", dataa)
			fmt.Println("after quiery agent")

			copy(sharedMemoryArray[:], result)

			// result, err := queryAgent("\\\\.\\pipe\\openssh-ssh-agent", append(lenBuf, buf...))
			// copyDataStruct *fs
			// fs = copyDataStruct (&cds)

			// cds := copyDataStruct{lParam}
			return 1
		}
	}
	// ret, handled := gohl.ProcNoDefault(hWnd, message, wParam, lParam)
	// if handled {
	// 	return uintptr(ret)
	// }
	// switch message {
	// case win.WM_CREATE:
	// 	println("win.WM_CREATE called", win.WM_CREATE)
	// }
	return win.DefWindowProc(hWnd, message, wParam, lParam)
}

func init() {
	// runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GOMAXPROCS(1)
}
