package win

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func AttachConsole() error {
	r1, _, err := attachConsole.Call(ATTACH_PARENT_PROCESS)
	if r1 == 0 {
		errno, ok := err.(syscall.Errno)
		if ok && errno == ERROR_INVALID_HANDLE {
			// console handle doesn't exist; not a real error, but the console handle will be invalid.
			return nil
		}
		return err
	}
	return nil
}

var oldStdout *os.File

func FixConsoleIfNeeded() error {
	// Keep old os.Stdout reference so it dont get GC'd and cleaned up
	// You never want to close file descriptors 0, 1, and 2.
	oldStdout = os.Stdout
	stdout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)

	var invalid syscall.Handle
	con := invalid

	if stdout == invalid {
		err := AttachConsole()
		if err != nil {
			return fmt.Errorf("attachconsole: %v", err)
		}
		if stdout == invalid {
			stdout, _ = syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
			con = stdout
		}
	}

	if con != invalid {
		// Make sure the console is configured to convert
		// \n to \r\n, like Go programs expect.
		h := windows.Handle(con)
		var st uint32
		err := windows.GetConsoleMode(h, &st)
		if err != nil {
			return fmt.Errorf("GetConsoleMode: %v", err)
		}
		err = windows.SetConsoleMode(h, st&^windows.DISABLE_NEWLINE_AUTO_RETURN)
		if err != nil {
			return fmt.Errorf("SetConsoleMode: %v", err)
		}
	}

	if stdout != invalid {
		os.Stdout = os.NewFile(uintptr(stdout), "stdout")
	}

	return nil
}
