package utils

import (
	"bufio"
	"errors"
	"fmt"
	"log"

	"github.com/Microsoft/go-winio"
)

func queryAgent(pipeName string, buf []byte) (result []byte, err error) {
	if len(buf) > agentMaxMessageLength {
		err = errors.New("Message too long")
		return
	}

	debug := false
	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to pipe %s: %w", pipeName, err)
	} else if debug {
		log.Printf("Connected to %s: %d", pipeName, len(buf))
	}
	defer conn.Close()

	l, err := conn.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot write to pipe %s: %w", pipeName, err)
	} else if debug {
		log.Printf("Sent to %s: %d", pipeName, l)
	}

	reader := bufio.NewReader(conn)
	res := make([]byte, agentMaxMessageLength)

	l, err = reader.Read(res)
	if err != nil {
		return nil, fmt.Errorf("cannot read from pipe %s: %w", pipeName, err)
	} else if debug {
		log.Printf("Received from %s: %d", pipeName, l)
	}
	return res[0:l], nil
}
