package protocol

import (
	"fmt"
)

// Type mirrors io.onedev.agent.MessageTypes enum ordinals.
type Type byte

const (
	TypeHeartBeat Type = iota
	TypeAgentData
	TypeError
	TypeUpdate
	TypeRestart
	TypeStop
	TypeUpdateAttributes
	TypeRequest
	TypeResponse
	TypeLog
	TypeCancelJob
	TypeStopWorkspace
	TypeReportJobWorkdir
	TypeResumeJob
	TypeJobShellOutput
	TypeJobShellOpen
	TypeJobShellTerminate
	TypeJobShellExit
	TypeWorkspaceShellOutput
	TypeWorkspaceShellOpen
	TypeWorkspaceShellTerminate
	TypeWorkspaceShellExit
	TypeWorkspaceShellInput
	TypeWorkspaceShellResize
	TypeJobShellInput
	TypeJobShellResize
	TypeDeleteWorkspace
)

// MaxMessageBytes matches io.onedev.agent.Agent.MAX_MESSAGE_BYTES.
const MaxMessageBytes = 64*1024*1024*4 + 100

// SocketIdleTimeoutMs matches io.onedev.agent.Agent.SOCKET_IDLE_TIMEOUT.
const SocketIdleTimeoutMs = 30000

// ServerPath is the websocket endpoint used by OneDev agents.
const ServerPath = "/~server"

// Message is the OneDev agent websocket frame: [type][payload...].
type Message struct {
	Type Type
	Data []byte
}

// Encode serializes a message for the OneDev websocket protocol.
func Encode(msg Message) ([]byte, error) {
	if len(msg.Data) > MaxMessageBytes-1 {
		return nil, fmt.Errorf("message payload exceeds max size")
	}
	out := make([]byte, len(msg.Data)+1)
	out[0] = byte(msg.Type)
	copy(out[1:], msg.Data)
	return out, nil
}

// Decode parses a websocket binary frame from the OneDev agent protocol.
func Decode(frame []byte) (Message, error) {
	if len(frame) == 0 {
		return Message{}, fmt.Errorf("empty frame")
	}
	if int(frame[0]) >= len(typeNames) {
		return Message{}, fmt.Errorf("unknown message type %d", frame[0])
	}
	data := make([]byte, len(frame)-1)
	copy(data, frame[1:])
	return Message{Type: Type(frame[0]), Data: data}, nil
}

// String returns the Java enum name for a message type.
func (t Type) String() string {
	if int(t) < len(typeNames) {
		return typeNames[t]
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}

var typeNames = []string{
	"HEART_BEAT",
	"AGENT_DATA",
	"ERROR",
	"UPDATE",
	"RESTART",
	"STOP",
	"UPDATE_ATTRIBUTES",
	"REQUEST",
	"RESPONSE",
	"LOG",
	"CANCEL_JOB",
	"STOP_WORKSPACE",
	"REPORT_JOB_WORKDIR",
	"RESUME_JOB",
	"JOB_SHELL_OUTPUT",
	"JOB_SHELL_OPEN",
	"JOB_SHELL_TERMINATE",
	"JOB_SHELL_EXIT",
	"WORKSPACE_SHELL_OUTPUT",
	"WORKSPACE_SHELL_OPEN",
	"WORKSPACE_SHELL_TERMINATE",
	"WORKSPACE_SHELL_EXIT",
	"WORKSPACE_SHELL_INPUT",
	"WORKSPACE_SHELL_RESIZE",
	"JOB_SHELL_INPUT",
	"JOB_SHELL_RESIZE",
	"DELETE_WORKSPACE",
}