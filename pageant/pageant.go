package pageant

type Pageant struct {
	sshPipe       string
	noPageantPipe bool
}

func New(sshPipe string, noPageantPipe bool) *Pageant {
	return &Pageant{
		sshPipe:       sshPipe,
		noPageantPipe: noPageantPipe,
	}
}
