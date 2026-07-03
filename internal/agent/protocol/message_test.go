package protocol

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := Message{Type: TypeUpdate, Data: []byte("16.0.2")}
	encoded, err := Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0] != byte(TypeUpdate) {
		t.Fatalf("type byte = %d, want %d", encoded[0], TypeUpdate)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Type != original.Type {
		t.Fatalf("type = %s, want %s", decoded.Type, original.Type)
	}
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Fatalf("data = %q, want %q", decoded.Data, original.Data)
	}
}