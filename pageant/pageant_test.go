package pageant

import (
	"reflect"
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
		}},
		{"an-ssh-pipe", args{[]Option{WithSSHPipe(`\\.\an-ssh-pipe\`)}}, &Pageant{
			sshPipe:     `\\.\an-ssh-pipe\`,
			pageantPipe: true,
		}},
		{"no-pageant-pipe", args{[]Option{WithPageantPipe(false)}}, &Pageant{
			sshPipe:     `\\.\pipe\openssh-ssh-agent`,
			pageantPipe: false,
		}},
		{"two-pipes", args{[]Option{WithSSHPipe(`\\.\an-ssh-pipe\`), WithPageantPipe(false)}}, &Pageant{
			sshPipe:     `\\.\an-ssh-pipe\`,
			pageantPipe: false,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
