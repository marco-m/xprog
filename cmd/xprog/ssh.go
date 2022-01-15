package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

type SshCmd struct {
	CommonArgs
	SshConfig string `arg:"--cfg,required" help:"path to a ssh_config file"`
	Sudo      bool   `help:"run the test binary with sudo"`
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
	log := self.opts.logger
	log.Debug("ssh", "testbinary:", self.TestBinary,
		"gotestflag:", self.GoTestFlag)
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
			ssh.PublicKeys(signer),
		},
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

	log.Debug("create scp session 1")
	scpClient, err := scp.NewClientBySSH(conn)
	if err != nil {
		return fmt.Errorf("sshRun: create scp session 1: %s", err)
	}

	dstTestBinary := "./" + path.Base(self.TestBinary)
	log.Debug("scp TestBinary host -> target",
		"src", self.TestBinary, "dst", dstTestBinary)
	fi, err := os.Open(self.TestBinary)
	if err != nil {
		return fmt.Errorf("sshRun: scp TestBinary: %s", err)
	}
	defer fi.Close()
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := scpClient.CopyFromFile(ctx, *fi, dstTestBinary, "0755"); err != nil {
			return fmt.Errorf("sshRun: scp copy TestBinary: %s", err)
		}
	}
	// If "go test -coverprofile", adapt accordingly
	var coverprofile, tgtCoverprofile string
	for i, flag := range self.GoTestFlag {
		tokens := strings.Split(flag, "=")
		if tokens[0] == "-test.coverprofile" {
			coverprofile = tokens[1]
			tgtCoverprofile = path.Base(coverprofile)
			self.GoTestFlag[i] = "-test.coverprofile=" + tgtCoverprofile
			break
		}
	}

	//
	// Execute the test binary.
	//
	log.Debug("create ssh session")
	sess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("sshRun: create ssh session: %s", err)
	}
	defer sess.Close()
	sess.Stdin = os.Stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	cmd := []string{"XPROG_SYS_TARGET=" + self.addr}
	if self.Sudo {
		cmd = append(cmd, "sudo", " --preserve-env=XPROG_SYS_TARGET")
	}
	cmd = append(cmd, dstTestBinary)
	cmd = append(cmd, self.GoTestFlag...)
	log.Debug("ssh execute TestBinary", "cmd", cmd)
	if err := sess.Run(strings.Join(cmd, " ")); err != nil {
		return fmt.Errorf("sshRun: execute TestBinary: %s", err)
	}

	// If no coverprofile, we are done.
	if coverprofile == "" {
		return nil
	}

	// Copy the coverprofile from target to host

	log.Debug("scp coverprofile target -> host",
		"src", tgtCoverprofile, "dst", coverprofile)
	fi, err = os.Create(coverprofile)
	if err != nil {
		return fmt.Errorf("sshRun: coverprofile: %s", err)
	}
	defer fi.Close()
	log.Debug("create scp session 2")
	scpClient2, err := scp.NewClientBySSH(conn)
	if err != nil {
		return fmt.Errorf("sshRun: create scp session 2: %s", err)
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := scpClient2.CopyFromRemote(ctx, fi, tgtCoverprofile); err != nil {
			return fmt.Errorf("sshRun: scp copy coverprofile: %s", err)
		}
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
