package sshserver

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func publicKeyCallback(authorized []ssh.PublicKey) func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error) {
	return func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if len(authorized) == 0 {
			return &ssh.Permissions{}, nil
		}
		for _, allowed := range authorized {
			if bytes.Equal(allowed.Marshal(), key.Marshal()) {
				return &ssh.Permissions{}, nil
			}
		}
		return nil, fmt.Errorf("public key not authorized")
	}
}