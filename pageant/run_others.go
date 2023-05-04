//go:build !windows
// +build !windows

package pageant

import (
	"errors"
	"log"
)

var defaultHandlerFunc = func(_ *Pageant, _ []byte) ([]byte, error) {
	return nil, errors.New("not supported")
}

func (p *Pageant) Run() {
	log.Fatalf("winssh-pageant bridge only supported on Windows")
}
