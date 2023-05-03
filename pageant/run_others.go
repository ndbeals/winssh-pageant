//go:build !windows
// +build !windows

package pageant

import (
	"log"
)

func (p *Pageant) Run() {
	log.Fatalf("winssh-pageant bridge only supported on Windows")
}
