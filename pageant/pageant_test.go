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
			sshPipe:     `\\.\pipe\openssh-ssh-agent`,
			pageantPipe: true,
			HandlerFunc: defaultHandlerFunc,
		}},
		{"an-ssh-pipe", args{[]Option{WithSSHPipe(`\\.\an-ssh-pipe\`)}}, &Pageant{
			sshPipe:     `\\.\an-ssh-pipe\`,
			pageantPipe: true,
			HandlerFunc: defaultHandlerFunc,
		}},
		{"no-pageant-pipe", args{[]Option{WithPageantPipe(false)}}, &Pageant{
			sshPipe:     `\\.\pipe\openssh-ssh-agent`,
			pageantPipe: false,
			HandlerFunc: defaultHandlerFunc,
		}},
		{"two-pipes", args{[]Option{WithSSHPipe(`\\.\an-ssh-pipe\`), WithPageantPipe(false)}}, &Pageant{
			sshPipe:     `\\.\an-ssh-pipe\`,
			pageantPipe: false,
			HandlerFunc: defaultHandlerFunc,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.opts...)
			if got.sshPipe != tt.want.sshPipe {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
			if got.pageantPipe != tt.want.pageantPipe {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
			if got.HandlerFunc == nil {
				t.Errorf("got.HandlerFunc is nil")
			}
		})
	}
}
