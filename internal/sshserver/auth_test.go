package sshserver

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/euxaristia/twodev/internal/config"
	"golang.org/x/crypto/ssh"
)

func TestPublicKeyCallback(t *testing.T) {
	allowed := mustSigner(t)
	other := mustSigner(t)

	callback := publicKeyCallback(config.SSHAuthorizedKeys{
		Enforce: true,
		Keys:    []ssh.PublicKey{allowed.PublicKey()},
	})
	if _, err := callback(nil, allowed.PublicKey()); err != nil {
		t.Fatalf("allowed key rejected: %v", err)
	}
	if _, err := callback(nil, other.PublicKey()); err == nil {
		t.Fatal("expected unauthorized key to be rejected")
	}
}

func TestPublicKeyCallbackAllowsAllWhenNotEnforced(t *testing.T) {
	pub := mustSigner(t).PublicKey()
	callback := publicKeyCallback(config.SSHAuthorizedKeys{})
	if _, err := callback(nil, pub); err != nil {
		t.Fatalf("expected open auth, got %v", err)
	}
}

func TestPublicKeyCallbackDeniesWhenEnforcedAndEmpty(t *testing.T) {
	pub := mustSigner(t).PublicKey()
	callback := publicKeyCallback(config.SSHAuthorizedKeys{Enforce: true})
	if _, err := callback(nil, pub); err == nil {
		t.Fatal("expected deny when authorized_keys exists but is empty")
	}
}

func mustSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}