package pageant

type Pageant struct {
	sshPipe     string
	pageantPipe bool
}

func New(opts ...Option) *Pageant {
	p := &Pageant{
		sshPipe:     `\\.\pipe\openssh-ssh-agent`,
		pageantPipe: true,
	}
	for _, applyTo := range opts {
		applyTo(p)
	}
	return p
}
