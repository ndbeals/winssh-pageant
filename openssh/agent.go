package openssh

import (
	"encoding/binary"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/rs/zerolog/log"
)

// AgentMaxMessageLength is the maximum length of a message sent to the agent
const (
	AgentMaxMessageLength      = 1<<14 - 1 // 16383
	SSH_AGENT_FAIL        byte = 0x05
)

var genericFail = []byte{0x00, 0x00, 0x00, 0x01, SSH_AGENT_FAIL}

// QueryAgent provides a way to query the named windows openssh agent pipe
func QueryAgent(pipeName string, buf []byte) (result []byte, err error) {
	// log := log.With().Str("pipe_name", pipeName).Logger()

	// l.Debug().Msgf("QueryAgent: %s", hex.EncodeToString(buf))
	log.Debug().Str("pipe_name", pipeName).Msg("Querying OpenSSH Agent")

	if len(buf) > AgentMaxMessageLength {
		log.Error().Msgf("message of lenth: %d is longer than max: %d", len(buf), AgentMaxMessageLength)
		// fmt.Println("message too long")
		return genericFail, nil
	}

	log.Debug().Msg("Dialing agent pipe")
	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("cannot connect to pipe: %s", pipeName)
		// fmt.Printf("cannot connect to pipe %s: %s\n", pipeName, err.Error())
		return genericFail, nil
	}
	defer conn.Close()

	log.Debug().Msg("Dialed successfully")

	// If the agent needs the user to do something, give them time to do so, but don't wait forever.
	err = conn.SetDeadline(time.Now().Add(time.Second * 60))
	if err != nil {
		log.Error().Stack().Err(err).Msgf("cannot set deadline on pipe: %s", pipeName)
	}

	log.Debug().Msg("Writing putty request to openssh pipe")

	_, err = conn.Write(buf)
	if err != nil {
		log.Error().Stack().Err(err).Msg("cannot write to openssh pipe")
		// fmt.Printf("cannot write ssh client request to agent pipe %s: %s\n", pipeName, err.Error())
		return genericFail, nil
	}

	log.Debug().Msg("request wrote successfully, reading response")

	err = conn.SetDeadline(time.Now().Add(time.Second * 60)) // Update deadline
	if err != nil {
		log.Error().Stack().Err(err).Msgf("cannot set deadline on pipe: %s", pipeName)
	}
	// <https://github.com/openssh/openssh-portable/blob/4e636cf/PROTOCOL.agent>
	// first 4 bytes are messageSizeBuf uint32
	messageSizeBuf := make([]byte, 4)
	_, err = conn.Read(messageSizeBuf)
	if err != nil {
		log.Error().Stack().Err(err).Msg("cannot read response message size from openssh pipe")
		return genericFail, nil
	}
	messageSize := binary.BigEndian.Uint32(messageSizeBuf)

	log.Debug().Msgf("message size: %d", messageSize)

	// next byte is the reply type code
	replyCode := make([]byte, 1)
	_, err = conn.Read(replyCode)
	if err != nil {
		log.Error().Stack().Err(err).Msg("cannot read response message type from openssh pipe")
		return append(messageSizeBuf, SSH_AGENT_FAIL), nil
	}
	log.Debug().Msgf("reply code: %d", replyCode[0])
	if replyCode[0] == SSH_AGENT_FAIL {
		return append(messageSizeBuf, replyCode...), nil
	}

	// https://datatracker.ietf.org/doc/html/draft-miller-ssh-agent-04#section-3
	messageContents := make([]byte, messageSize-1)
	_, err = conn.Read(messageContents)
	if err != nil {
		log.Error().Stack().Err(err).Msg("cannot read response message contents from openssh pipe")
		return append(messageSizeBuf, SSH_AGENT_FAIL), nil
	}
	log.Debug().Msgf("message contents length: %d", len(messageContents))

	concatResults := append(messageSizeBuf, replyCode...)
	concatResults = append(concatResults, messageContents...)

	log.Debug().Msgf("response size: %d", len(concatResults))

	return concatResults, nil
}
