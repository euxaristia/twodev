package protocol

import (
	"encoding/json"
	"fmt"
)

// Call is a JSON RPC envelope carried in REQUEST/RESPONSE frames.
type Call struct {
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

const MethodRunJob = "runJob"

// RunJobPayload describes a build job to execute on an agent.
type RunJobPayload struct {
	Token       string `json:"token"`
	ProjectID   int64  `json:"projectId"`
	ProjectPath string `json:"projectPath"`
	JobName     string `json:"jobName"`
	BuildNumber int    `json:"buildNumber"`
	BuildSpec   string `json:"buildSpec"`
	Branch      string `json:"branch"`
	CommitHash  string `json:"commitHash"`
	RepoRoot    string `json:"repoRoot"`
	CloneURL    string `json:"cloneUrl,omitempty"`
}

// RunJobResult reports agent-side job completion.
type RunJobResult struct {
	OK bool `json:"ok"`
}

// EncodeCall serializes a call for the websocket REQUEST/RESPONSE payload.
func EncodeCall(call Call) ([]byte, error) {
	return json.Marshal(call)
}

// DecodeCall parses a call from websocket payload bytes.
func DecodeCall(data []byte) (Call, error) {
	var call Call
	if err := json.Unmarshal(data, &call); err != nil {
		return Call{}, fmt.Errorf("decode call: %w", err)
	}
	if call.ID == "" {
		return Call{}, fmt.Errorf("call id is required")
	}
	return call, nil
}