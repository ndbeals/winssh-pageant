package sshagent

import (
	"bufio"
	"fmt"

	"github.com/Microsoft/go-winio"
)

// QueryAgent provides a way to query the named windows openssh agent pipe
func QueryAgent(pipeName string, buf []byte, agentMaxMessageLength int) (result []byte, err error) {
	if len(buf) > agentMaxMessageLength {
		return nil, fmt.Errorf("Message too long")
	}

	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to pipe %s: %w", pipeName, err)
	}
	defer conn.Close()

	l, err := conn.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot write to pipe %s: %w", pipeName, err)
	}

	reader := bufio.NewReader(conn)
	res := make([]byte, agentMaxMessageLength)

	l, err = reader.Read(res)
	if err != nil {
		return nil, fmt.Errorf("cannot read from pipe %s: %w", pipeName, err)
	}

	return res[0:l], nil
}
