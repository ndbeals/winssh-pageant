package sshagent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/Microsoft/go-winio"
)

// AgentMaxMessageLength is the maximum length of a message sent to the agent
const (
	AgentMaxMessageLength      = 1<<14 - 1 // 16383
	SSH_AGENT_FAIL        byte = 0x05
)

var genericFail = []byte{0x00, 0x00, 0x00, 0x01, SSH_AGENT_FAIL}

// QueryAgent provides a way to query the named windows openssh agent pipe
func QueryAgent(pipeName string, buf []byte) (result []byte, err error) {
	if len(buf) > AgentMaxMessageLength {
		fmt.Println("message too long")
		return genericFail, nil
	}

	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		fmt.Printf("cannot connect to pipe %s: %s\n", pipeName, err.Error())
		return genericFail, nil
	}
	defer conn.Close()
	// If the agent needs the user to do something, give them time to do so, but don't wait forever.
	conn.SetDeadline(time.Now().Add(time.Second * 2))

	_, err = conn.Write(buf)
	if err != nil {
		fmt.Printf("cannot write ssh client request to agent pipe %s: %s\n", pipeName, err.Error())
		return genericFail, nil
	}

	conn.SetDeadline(time.Now().Add(time.Second * 60)) // Update deadline
	// <https://github.com/openssh/openssh-portable/blob/4e636cf/PROTOCOL.agent>
	// first 4 bytes are messageSizeBuf uint32
	messageSizeBuf := make([]byte, 4)
	_, err = conn.Read(messageSizeBuf)
	if err != nil {
		switch {
		case errors.Is(err, winio.ErrTimeout):
			fmt.Printf("Timeout waiting for user input %s: %s\n", pipeName, err.Error())
		default:
			fmt.Printf("Cannot read message size from pipe %s: %s\n", pipeName, err.Error())
		}
		return genericFail, nil
	}
	messageSize := binary.BigEndian.Uint32(messageSizeBuf)

	// next byte is the reply type code
	replyCode := make([]byte, 1)
	_, err = conn.Read(replyCode)
	if err != nil {
		fmt.Printf("Cannot read message type from pipe %s: %s\n", pipeName, err.Error())
		return append(messageSizeBuf, SSH_AGENT_FAIL), nil
	}
	if replyCode[0] == SSH_AGENT_FAIL {
		return append(messageSizeBuf, replyCode...), nil
	}

	// https://datatracker.ietf.org/doc/html/draft-miller-ssh-agent-04#section-3
	messageContents := make([]byte, messageSize-1)
	_, err = conn.Read(messageContents)
	if err != nil {
		fmt.Printf("cannot read message contents from pipe %s: %s\n", pipeName, err.Error())
		return append(messageSizeBuf, SSH_AGENT_FAIL), nil
	}

	concatResults := append(messageSizeBuf, replyCode...)
	concatResults = append(concatResults, messageContents...)

	return concatResults, nil
}
