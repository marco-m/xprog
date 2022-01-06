package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type SshCmd struct {
	CommonArgs
	SshConfig string `arg:"required" help:"path to a ssh_config file"`
}

func (self SshCmd) Run(root Root) error {
	rd, err := os.Open(self.SshConfig)
	if err != nil {
		return fmt.Errorf("ssh_config: %s", err)
	}
	defer rd.Close()
	sshConf, err := parseSshConfig(rd)
	if err != nil {
		return err
	}
	// Currently we always take the first Host block.
	host := sshConf[0]

	privateKeyPath, err := host.Get("IdentityFile")
	if err != nil {
		return err
	}

	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %s", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %s", err)
	}

	user, err := host.Get("User")
	if err != nil {
		return err
	}

	strictHostKeyChecking, err := host.Get("StrictHostKeyChecking")
	if err != nil {
		return err
	}
	if strictHostKeyChecking != "no" {
		return fmt.Errorf("StrictHostKeyChecking=%s but we support only 'no'", strictHostKeyChecking)
	}

	sshCfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	hostName, err := host.Get("HostName")
	if err != nil {
		return err
	}
	port, err := host.Get("Port")
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", hostName, port)
	conn, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return fmt.Errorf("ssh: %s", err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("ssh session: %s", err)
	}
	defer sess.Close()
	sess.Stdin = os.Stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	scpSess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("scp session: %s", err)
	}
	defer scpSess.Close()

	dstTestBinary := "./" + path.Base(self.TestBinary)
	err = scp.CopyPath(self.TestBinary, dstTestBinary, scpSess)
	if err != nil {
		return fmt.Errorf("scp copy: %s", err)
	}

	cmd := []string{dstTestBinary}
	cmd = append(cmd, self.GoTestFlag...)
	return sess.Run(strings.Join(cmd, " "))
}

// Host is a `Host` block in a ssh_config file.
type Host map[string]string

// Get returns self[key] if found or error if not found.
func (self Host) Get(key string) (string, error) {
	val := self[key]
	if val != "" {
		return val, nil
	}
	return "", fmt.Errorf("ssh_config: missing key %s", key)
}

// GetDef returns self[key] if found or def if not found.
func (self Host) GetDef(key string, def string) string {
	val := self[key]
	if val != "" {
		return val
	}
	return def
}

// parseSshConfig is a simplistic and partial parser of ssh_config files.
// It knows only about `Host` blocks.
func parseSshConfig(rd io.Reader) ([]Host, error) {
	var hosts []Host
	host := Host{}
	scanner := bufio.NewScanner(rd)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		if have, want := len(tokens), 2; have != want {
			return nil, fmt.Errorf("parseSshConfig: line '%s': %d tokens instead of %d",
				line, have, want)
		}
		k, v := tokens[0], tokens[1]

		if len(host) == 0 && k != "Host" {
			return nil,
				fmt.Errorf("parseSshConfig: line '%s': block must begin with 'Host'",
					line)
		}

		if len(host) > 0 && k == "Host" {
			// beginning of new block

			hosts = append(hosts, host)
			host = Host{}
		}

		if old, ok := host[k]; ok {
			return nil,
				fmt.Errorf("parseSshConfig: block 'Host %s': duplicated k/v: '%s %s', previous: '%s %s'",
					host["Host"], k, v, k, old)
		}

		host[k] = v
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(host) == 1 {
		return nil, fmt.Errorf("parseSshConfig: empty 'Host' block: %v", host)
	}

	if len(host) > 0 {
		hosts = append(hosts, host)
	}

	if len(hosts) == 0 {
		return nil, errors.New("parseSshConfig: empty file")
	}

	seen := map[string]bool{}
	for _, host := range hosts {
		h := host["Host"]
		if seen[h] {
			return nil, fmt.Errorf("parseSshConfig: duplicated block 'Host %s'", h)
		}
		seen[h] = true
	}

	return hosts, nil
}
