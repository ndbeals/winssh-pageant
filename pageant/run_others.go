//go:build !windows
// +build !windows

package pageant

import (
	"fmt"
	"os"
)

func (p *Pageant) Run() {
	fmt.Println("winssh-pageant bridge only supported on Windows")
	os.Exit(1)
}
