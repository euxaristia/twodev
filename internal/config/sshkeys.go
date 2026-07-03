package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// LoadSSHAuthorizedKeys reads public keys from site/conf/authorized_keys.
// Returns nil when the file is missing so the SSH server can allow all keys in dev.
func LoadSSHAuthorizedKeys(siteDir string) ([]ssh.PublicKey, error) {
	path := filepath.Join(siteDir, "conf", "authorized_keys")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := checkTokenFilePermissions(path); err != nil {
		_ = file.Close()
		return nil, err
	}
	defer file.Close()

	var keys []ssh.PublicKey
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
		if err != nil {
			return nil, err
		}
		keys = append(keys, pub)
	}
	return keys, scanner.Err()
}