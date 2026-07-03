package protocol

import (
	"encoding/json"
	"testing"
)

func TestCallRoundTrip(t *testing.T) {
	payload, _ := json.Marshal(RunJobPayload{Token: "t1", JobName: "CI"})
	original := Call{ID: "abc", Method: MethodRunJob, Payload: payload}
	data, err := EncodeCall(original)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCall(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.ID != original.ID || decoded.Method != original.Method {
		t.Fatalf("call = %+v, want %+v", decoded, original)
	}
}