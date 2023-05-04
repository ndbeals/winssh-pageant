package pageant

type Pageant struct {
	sshPipe     string
	pageantPipe bool
	HandlerFunc func(p *Pageant, result []byte) ([]byte, error)
}

func New(opts ...Option) *Pageant {
	// initialize with defaults
	p := &Pageant{
		sshPipe:     `\\.\pipe\openssh-ssh-agent`,
		pageantPipe: true,
	}
	// set the default handler func
	p.HandlerFunc = defaultHandlerFunc
	// apply options
	for _, applyTo := range opts {
		applyTo(p)
	}
	return p
}
