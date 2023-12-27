package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"github.com/ndbeals/winssh-pageant/internal/win"
	"github.com/ndbeals/winssh-pageant/pageant"
)

var (
	sshPipe       = flag.String("sshpipe", `\\.\pipe\openssh-ssh-agent`, "Named pipe for Windows OpenSSH agent")
	noPageantPipe = flag.Bool("no-pageant-pipe", false, "Toggle pageant named pipe proxying (this is different from the windows OpenSSH pipe)")
)

func main() {
	verbose := flag.BoolP("verbose", "v", false, "Turn on verbose logging")
	_ = verbose
	flag.Parse()

	if *verbose || (isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
		err := win.FixConsoleIfNeeded()
		if err != nil {
			log.Error().Stack().Err(err).Msg("FixConsole failed")
		}
		configureLogger(*verbose)
	} else {
		log.Logger = zerolog.Nop()
	}

	log.Info().Msg("Starting winssh-pageant")

	p := pageant.NewDefaultHandler(*sshPipe, !*noPageantPipe)

	p.Run()
}

func configureLogger(verbose bool) {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	output := zerolog.NewConsoleWriter()
	output.TimeFormat = "2006-01-02 15:04:05"
	log.Logger = log.Output(output)

	if !verbose {
		return
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	output.FieldsExclude = []string{zerolog.ErrorStackFieldName}
	output.FormatExtra = func(evt map[string]interface{}, buf *bytes.Buffer) error {
		stack, ok := evt[zerolog.ErrorStackFieldName].(string)
		if ok {
			_, err := buf.WriteString(stack)
			if err != nil {
				return errors.Wrap(err, "cannot write stack to buffer")
			}
		}
		return nil
	}
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		type stackTracer interface {
			StackTrace() errors.StackTrace
		}
		var sterr stackTracer
		for err != nil {
			possible, ok := err.(stackTracer)
			if ok {
				sterr = possible
			}
			err = errors.Unwrap(err)
		}
		if sterr == nil {
			return nil
		}

		st := sterr.StackTrace()
		return fmt.Sprintf("%+v", st[:len(st)-1])
	}

	log.Logger = log.Output(output)
}
