package main

import (
	utils "github.com/ndbeals/winssh-pageant/internal"

	"github.com/lxn/win"
)

func main() {

	inst := win.GetModuleHandle(nil)
	utils.WinMain(inst)
}
