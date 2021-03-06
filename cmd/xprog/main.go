package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/alexflint/go-arg"
	"github.com/hashicorp/go-hclog"
)

// Must be filled by the linker via build script until Go 1.18
var version = "(devel)" // to match the default from runtime/debug

type Opts struct {
	Verbose bool `arg:"-v,--verbose" help:"verbosity level"`
	//
	Help   *HelpCmd   `arg:"subcommand:help" help:"display extensive help"`
	Direct *DirectCmd `arg:"subcommand:direct" help:"run the test binary directly on the host"`
	Ssh    *SshCmd    `arg:"subcommand:ssh" help:"upload and run the test binary on SSH target"`
	// proposed new API for go-arg:
	// Extra   []string `arg:"end-of-options"`
	// instead of:
	// Extra []string `arg:"positional"`
	//
	out    io.Writer
	logger hclog.Logger
}

type CommonArgs struct {
	TestBinary string   `arg:"required,positional" help:"path to the test binary created by go test"`
	GoTestFlag []string `arg:"positional" help:"flags for go test; put '-- ' before the first one (to signal end of options)"`
}

type DirectCmd struct {
	CommonArgs
	opts Opts
}

type HelpCmd struct{}

const help = `xprog -- a test runner for go test -exec.

Generic usage from go test:

    go test -exec="xprog <command> [opts] --" <go-packages> [go-test-flags]

Cross-compile the tests and run them on the target OS, connect via SSH:

    GOOS=linux go test -exec="xprog ssh [opts] --" <go-packages> [go-test-flags]

As above, but collect also code coverage and show it on the host:

    GOOS=linux go test -coverprofile=coverage.out -exec="xprog ssh [opts] --" <go-packages> [go-test-flags] &&
    go tool cover -html=coverage.out

To see xprog output, pass -v both to xprog and go test:

    go test -v -exec="xprog -v <command> [opts] --" <go-packages> [go-test-flags]
`

func main() {
	os.Exit(mainInt(os.Stderr, os.Args[1:]))
}

func mainInt(out io.Writer, args []string) int {
	var opts Opts
	err := parse(out, args, arg.Config{}, &opts)
	if err == parseOK {
		return 0
	}
	if err != nil {
		fmt.Fprintln(out, "xprog:", err)
		return 1
	}

	opts.out = out
	opts.logger = hclog.New(&hclog.LoggerOptions{
		Name:   "xprog",
		Output: out,
	})
	if opts.Verbose {
		opts.logger.SetLevel(hclog.Debug)
	}

	if err := runCommand(opts); err != nil {
		fmt.Fprintln(out, "xprog:", err)
		return 1
	}

	return 0
}

var parseOK = errors.New("parse OK")

func parse(out io.Writer, args []string, config arg.Config, dests ...interface{}) error {
	parser, err := arg.NewParser(config, dests...)
	if err != nil {
		return err
	}

	err = parser.Parse(args)
	switch {
	case err == arg.ErrHelp:
		parser.WriteHelp(out)
		return parseOK
	case err == arg.ErrVersion:
		// I am not sure this actually works???
		// Courtesy of https://github.com/burrowers/garble/pull/220, waiting for Go 1.18
		// don't overwrite the version if it was set by -ldflags=-X
		if info, ok := debug.ReadBuildInfo(); ok && version == "(devel)" {
			mod := &info.Main
			if mod.Replace != nil {
				mod = mod.Replace
			}
			version = mod.Version
		}
		fmt.Fprintln(out, "xprog version", version)
		return parseOK
	case err != nil:
		parser.WriteUsage(out)
		return err
	}

	// go-arg allows to invoke the program without subcommands. Since it would
	// not make sense for us, we check ourselves.
	if parser.Subcommand() == nil {
		parser.WriteUsage(out)
		return fmt.Errorf("missing subcommand")
	}

	return nil
}

func runCommand(opts Opts) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %s", err)
	}
	opts.logger.Debug("starting", "cwd", cwd)
	defer opts.logger.Debug("terminating")

	switch {
	case opts.Help != nil:
		return opts.Help.Run(opts)
	case opts.Direct != nil:
		return opts.Direct.Run(opts)
	case opts.Ssh != nil:
		return opts.Ssh.Run(opts)
	default:
		return fmt.Errorf("unwired command")
	}
}

func (self HelpCmd) Run(opts Opts) error {
	fmt.Fprintln(opts.out, help)
	return nil
}

func (self DirectCmd) Run(opts Opts) error {
	self.opts = opts
	self.opts.logger.Debug("direct", "testbinary:", self.TestBinary,
		"gotestflag:", self.GoTestFlag)

	cmd := exec.Command(self.TestBinary, self.GoTestFlag...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("direct: %s", err)
	}
	return nil
}
