package pageant

import "log"

type PageantRequestHandler func(p *Pageant, result []byte) ([]byte, error)

type Pageant struct {
	SSHAgentPipe string // pipe for windows openssh agent (e.g \\.\pipe\openssh-ssh-agent)
	pageantPipe  bool   // enable pageant named pipe proxying (not the same as the windows openssh pipe)

	// This function is called when an incoming pageant key request is received.
	// The result of the function is sent back to the requesting client.
	PageantRequestHandler PageantRequestHandler
}

// New creates a new pageant with explicit arguments
func New(openSSHPipe string, enablePageantPipe bool, pageantRequestHandler PageantRequestHandler) *Pageant {
	return &Pageant{
		SSHAgentPipe:          openSSHPipe,
		pageantPipe:           enablePageantPipe,
		PageantRequestHandler: pageantRequestHandler,
	}
}

// NewDefaultHandler creates a new pageant with the default handler func
func NewDefaultHandler(openSSHPipe string, enablePageantPipe bool) *Pageant {
	return New(openSSHPipe, enablePageantPipe, defaultHandlerFunc)
}

// Configure the pageant with the given options if provided, otherwise use defaults
func NewWithOptions(opts ...Option) *Pageant {
	// initialize with defaults
	p := New(`\\.\pipe\openssh-ssh-agent`, true, defaultHandlerFunc)

	// apply options
	for _, applyTo := range opts {
		err := applyTo(p)
		if err != nil {
			log.Printf("Error applying option: %v\n", err)
		}
	}
	return p
}

type Option func(p *Pageant) error

func WithSSHPipe(sshPipe string) Option {
	return func(p *Pageant) error {
		p.SSHAgentPipe = sshPipe
		return nil
	}
}

func WithPageantPipe(pageantPipe bool) Option {
	return func(p *Pageant) error {
		p.pageantPipe = pageantPipe
		return nil
	}
}

func WithPageantRequestHandler(handler PageantRequestHandler) Option {
	return func(p *Pageant) error {
		p.PageantRequestHandler = handler
		return nil
	}
}
