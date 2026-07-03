package git

import (
	"context"
	"testing"
)

func TestVersion(t *testing.T) {
	svc := NewService("git")
	version, err := svc.Version(context.Background())
	if err != nil {
		t.Skip("git not available:", err)
	}
	if version == "" {
		t.Fatal("expected git version")
	}
}