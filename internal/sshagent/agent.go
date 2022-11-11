package sshagent

import (
	"bufio"
	"encoding/binary"
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
		fmt.Printf("cannot connect to pipe %s: %s", pipeName, err.Error())
		return genericFail, nil
	}
	defer conn.Close()
	// If the agent needs the user to do something, give them time to do so, but don't wait forever.
	conn.SetDeadline(time.Now().Add(time.Second * 20))

	_, err = conn.Write(buf)
	if err != nil {
		fmt.Printf("cannot write to pipe %s: %s", pipeName, err.Error())
		return genericFail, nil
	}

	// The buffer needs to be at least as large as the expected message size
	reader := bufio.NewReaderSize(conn, AgentMaxMessageLength)

	// Magic numbers from the ssh-agent protocol specification.
	// <https://github.com/openssh/openssh-portable/blob/4e636cf/PROTOCOL.agent>
	// first 4 bytes are magic numbers related to the named pipe
	magic := make([]byte, 4)
	_, err = reader.Read(magic)
	if err != nil {
		fmt.Printf("cannot read from pipe %s: %s", pipeName, err.Error())
		return genericFail, nil
	}
	// next byte is the reply code
	replyCode := make([]byte, 1)
	_, err = reader.Read(replyCode)
	if err != nil {
		fmt.Printf("cannot read from pipe %s: %s", pipeName, err.Error())
		return append(magic, []byte{SSH_AGENT_FAIL}...), nil
	}
	if replyCode[0] == SSH_AGENT_FAIL {
		return append(magic, replyCode...), nil
	}
	// next 4 bytes (Uint32) is the number of keys
	keyCountSlice := make([]byte, 4)
	_, err = reader.Read(keyCountSlice)
	if err != nil {
		fmt.Printf("cannot read from pipe %s: %s", pipeName, err.Error())
		return append(magic, []byte{SSH_AGENT_FAIL}...), nil
	}
	// convert to Uint32
	keyCount := binary.BigEndian.Uint32(keyCountSlice)

	// set to max agent message length minus the previous 9 bytes
	res := make([]byte, AgentMaxMessageLength-9)
	// verify the key count is > 0, otherwise skip
	byteCount := 0
	if keyCount > 0 {
		byteCount, err = reader.Read(res)
		if err != nil {
			fmt.Printf("cannot read from pipe %s: %s", pipeName, err.Error())
			return append(magic, []byte{SSH_AGENT_FAIL}...), nil
		}
	}

	// Concat all slices together
	concatRes := append(magic, replyCode...)
	concatRes = append(concatRes, keyCountSlice...)
	concatRes = append(concatRes, res[0:byteCount]...)

	res = nil // Explicitly clear the result to prevent memory leak
	return concatRes, nil
}
