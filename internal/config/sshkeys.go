package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh"
)

// SSHAuthorizedKeys holds parsed SSH public keys and whether enforcement is active.
type SSHAuthorizedKeys struct {
	Keys    []ssh.PublicKey
	Enforce bool
}

// LoadSSHAuthorizedKeys reads public keys from site/conf/authorized_keys.
// When the file is missing, Enforce is false and all keys are allowed.
// When the file exists but has no keys, Enforce is true and all connections are denied.
func LoadSSHAuthorizedKeys(siteDir string) (SSHAuthorizedKeys, error) {
	path := filepath.Join(siteDir, "conf", "authorized_keys")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return SSHAuthorizedKeys{}, nil
	}
	if err != nil {
		return SSHAuthorizedKeys{}, err
	}
	if err := checkSSHKeysFilePermissions(path); err != nil {
		_ = file.Close()
		return SSHAuthorizedKeys{}, err
	}
	defer file.Close()

	keys := make([]ssh.PublicKey, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
		if err != nil {
			return SSHAuthorizedKeys{}, err
		}
		keys = append(keys, pub)
	}
	if err := scanner.Err(); err != nil {
		return SSHAuthorizedKeys{}, err
	}
	return SSHAuthorizedKeys{Keys: keys, Enforce: true}, nil
}

func checkSSHKeysFilePermissions(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o002 != 0 {
		return fmt.Errorf("%s must not be world-writable", path)
	}
	return nil
}