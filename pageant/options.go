package pageant

type Option func(p *Pageant) error

func WithSSHPipe(sshPipe string) Option {
	return func(p *Pageant) error {
		p.sshPipe = sshPipe
		return nil
	}
}

func WithPageantPipe(pageantPipe bool) Option {
	return func(p *Pageant) error {
		p.pageantPipe = pageantPipe
		return nil
	}
}
