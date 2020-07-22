package main

import (
	"fmt"

	utils "githubcom/ndbeals/winssh-pageant/internal"

	"github.com/lxn/win"
)

func main() {

	fmt.Println("HETS")
	// fmt.Println(utils.AgentCopyDataID)
	inst := win.GetModuleHandle(nil)
	utils.WinMain(inst)
}
