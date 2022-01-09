package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type SshCmd struct {
	CommonArgs
	SshConfig string `arg:"--cfg,required" help:"path to a ssh_config file"`
	//
	opts   Opts
	sshCfg ssh.ClientConfig
	addr   string
}

func (self SshCmd) Run(opts Opts) error {
	self.opts = opts
	if err := self.prepare(); err != nil {
		return err
	}
	return self.execute()
}

func (self *SshCmd) prepare() error {
	rd, err := os.Open(self.SshConfig)
	if err != nil {
		return fmt.Errorf("sshRun: ssh_config: %s", err)
	}
	defer rd.Close()
	sshConf, err := parseSshConfig(rd)
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}
	// Currently we always take the first Host block.
	host := sshConf[0]

	privateKeyPath, err := host.Get("IdentityFile")
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}

	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("sshRun: unable to read private key: %s", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("sshRun: unable to parse private key: %s", err)
	}

	user, err := host.Get("User")
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}

	strictHostKeyChecking, err := host.Get("StrictHostKeyChecking")
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}
	if strictHostKeyChecking != "no" {
		return fmt.Errorf("sshRun: StrictHostKeyChecking=%s but we support only 'no'",
			strictHostKeyChecking)
	}

	self.sshCfg = ssh.ClientConfig{
		Timeout: 1 * time.Second,
		User:    user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	hostName, err := host.Get("HostName")
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}
	port, err := host.Get("Port")
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}

	self.addr = fmt.Sprintf("%s:%s", hostName, port)

	return nil
}

func (self SshCmd) execute() error {
	log := self.opts.logger

	log.Debug("ssh.Dial", "addr", self.addr)
	conn, err := ssh.Dial("tcp", self.addr, &self.sshCfg)
	if err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}
	defer conn.Close()

	log.Debug("create ssh session")
	sess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("sshRun: ssh session: %s", err)
	}
	defer sess.Close()
	sess.Stdin = os.Stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	log.Debug("create scp session")
	scpSess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("sshRun: scp session: %s", err)
	}
	defer scpSess.Close()

	dstTestBinary := "./" + path.Base(self.TestBinary)
	log.Debug("scp", "src", self.TestBinary, "dst", dstTestBinary)
	err = scp.CopyPath(self.TestBinary, dstTestBinary, scpSess)
	if err != nil {
		return fmt.Errorf("sshRun: scp copy: %s", err)
	}

	cmd := []string{dstTestBinary}
	cmd = append(cmd, self.GoTestFlag...)
	log.Debug("ssh execute", "cmd", cmd)
	if err := sess.Run(strings.Join(cmd, " ")); err != nil {
		return fmt.Errorf("sshRun: %s", err)
	}
	return nil
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
