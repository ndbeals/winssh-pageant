package pageant

import (
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want *Pageant
	}{
		{"defaults", args{nil}, &Pageant{
			SSHAgentPipe:          `\\.\pipe\openssh-ssh-agent`,
			pageantPipe:           true,
			PageantRequestHandler: defaultHandlerFunc,
		}},
		{"an-ssh-pipe", args{[]Option{WithSSHPipe(`\\.\an-ssh-pipe\`)}}, &Pageant{
			SSHAgentPipe:          `\\.\an-ssh-pipe\`,
			pageantPipe:           true,
			PageantRequestHandler: defaultHandlerFunc,
		}},
		{"no-pageant-pipe", args{[]Option{WithPageantPipe(false)}}, &Pageant{
			SSHAgentPipe:          `\\.\pipe\openssh-ssh-agent`,
			pageantPipe:           false,
			PageantRequestHandler: defaultHandlerFunc,
		}},
		{"two-pipes", args{[]Option{WithSSHPipe(`\\.\an-ssh-pipe\`), WithPageantPipe(false)}}, &Pageant{
			SSHAgentPipe:          `\\.\an-ssh-pipe\`,
			pageantPipe:           false,
			PageantRequestHandler: defaultHandlerFunc,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWithOptions(tt.args.opts...)
			if got.SSHAgentPipe != tt.want.SSHAgentPipe {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
			if got.pageantPipe != tt.want.pageantPipe {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
			if got.PageantRequestHandler == nil {
				t.Errorf("got.HandlerFunc is nil")
			}
		})
	}
}
