package sshserver

import (
	"bytes"
	"fmt"

	"github.com/euxaristia/twodev/internal/config"
	"golang.org/x/crypto/ssh"
)

func publicKeyCallback(authorized config.SSHAuthorizedKeys) func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error) {
	return func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if !authorized.Enforce {
			return &ssh.Permissions{}, nil
		}
		for _, allowed := range authorized.Keys {
			if bytes.Equal(allowed.Marshal(), key.Marshal()) {
				return &ssh.Permissions{}, nil
			}
		}
		return nil, fmt.Errorf("public key not authorized")
	}
}