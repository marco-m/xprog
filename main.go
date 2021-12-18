package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/alexflint/go-arg"
)

type cli struct {
	Root
	Help        *HelpCmd        `arg:"subcommand:help" help:"display extensive help"`
	Passthrough *PassthroughCmd `arg:"subcommand:passthrough" help:"run the test binary directly on the host"`
	Ssh         *SshCmd         `arg:"subcommand:ssh" help:"upload and run test binary via ssh"`
	// proposed new API for go-arg:
	// Extra   []string `arg:"end-of-options"`
	// instead of:
	// Extra []string `arg:"positional"`
}

type Root struct {
	Verbose bool `arg:"-v,--verbose" help:"verbosity level"`
	//
	Out io.Writer `arg:"-"`
}

type Common struct {
	TestBinary string   `arg:"required,positional" help:"path to the test binary created by go test"`
	GoTestFlag []string `arg:"positional" help:"flags for go test; put '-- ' before the first (to signal end-of-options)"`
}

type SshCmd struct {
	Common
}

type PassthroughCmd struct {
	Common
}

type HelpCmd struct {
}

const help = `xprog -- a test runner for "go test -exec"

Generic usage from go test:

    go test -exec='xprog <command> [opts] --' <go-packages> [go-test-flags]

Cross-compile the tests and run them on the target OS, connect via SSH:

    GOOS=linux go test -exec='xprog ssh [opts] --' <go-packages> [go-test-flags]

Note: to see xprog output, pass -v both to xprog and go test:

    go test -v -exec='xprog -v <command> [opts] --' <go-packages> [go-test-flags]
`

func main() {
	args := cli{
		Root: Root{Out: os.Stderr},
	}
	arg.MustParse(&args)
	p := arg.MustParse(&args)
	if p.Subcommand() == nil {
		p.Fail("missing subcommand")
	}
	if err := main2(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main2(args cli) error {
	switch {
	case args.Help != nil:
		return args.Help.Run(args.Root)
	case args.Passthrough != nil:
		return args.Passthrough.Run(args.Root)
	case args.Ssh != nil:
		// return sshCmd()
		return fmt.Errorf("unimplemented")
	default:
		return fmt.Errorf("unwired command")
	}
}

func (self HelpCmd) Run(root Root) error {
	fmt.Fprintln(root.Out, help)
	return nil
}

func (self PassthroughCmd) Run(root Root) error {
	if root.Verbose {
		fmt.Fprintln(root.Out, "Run:", self.TestBinary, self.GoTestFlag)
	}
	cmd := exec.Command(self.TestBinary, self.GoTestFlag...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
