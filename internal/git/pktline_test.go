package git

import (
	"bytes"
	"testing"
)

func TestAdvertiseService(t *testing.T) {
	out, err := AdvertiseService("git-upload-pack", []byte("refs"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("# service=git-upload-pack")) {
		t.Fatalf("unexpected output: %q", out)
	}
	if !bytes.Contains(out, []byte("refs")) {
		t.Fatal("expected git output appended")
	}
}