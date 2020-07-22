package main

import (
	utils "githubcom/ndbeals/winssh-pageant/internal"

	"github.com/lxn/win"
)

func main() {

	inst := win.GetModuleHandle(nil)
	utils.WinMain(inst)
}
