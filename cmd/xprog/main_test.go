package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdlineParsing(t *testing.T) {
	testCases := []struct {
		name     string
		cmdline  string
		wantCode int
		wantOut  string
	}{
		{
			name:     "missing subcommand",
			cmdline:  "",
			wantCode: 1,
			wantOut: `
Usage: xprog.test [--verbose] <command> [<args>]
xprog: missing subcommand
`,
		},
		{
			name:     "top level help",
			cmdline:  "-h",
			wantCode: 0,
			wantOut: `
Usage: xprog.test [--verbose] <command> [<args>]

Options:
  --verbose, -v          verbosity level
  --help, -h             display this help and exit

Commands:
  help                   display extensive help
  direct                 run the test binary directly on the host
  ssh                    upload and run the test binary on SSH target
`,
		},
		{
			name:     "help for direct",
			cmdline:  "direct -h",
			wantCode: 0,
			wantOut: `
Usage: xprog.test direct TESTBINARY [GOTESTFLAG [GOTESTFLAG ...]]

Positional arguments:
  TESTBINARY             path to the test binary created by go test
  GOTESTFLAG             flags for go test; put '-- ' before the first one (to signal end of options)

Global options:
  --verbose, -v          verbosity level
  --help, -h             display this help and exit
`,
		},
		{
			name:     "unknown command",
			cmdline:  "foo",
			wantCode: 1,
			wantOut: `
Usage: xprog.test [--verbose] <command> [<args>]
xprog: invalid subcommand: foo
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			{
				cmdline := strings.Fields(tc.cmdline)
				have := mainInt(&out, cmdline)
				want := tc.wantCode
				if have != want {
					t.Errorf("\nstatus code: have: %d; want: %d", have, want)
				}
			}

			{
				have := strings.Split(out.String(), "\n")
				want := strings.Split(tc.wantOut, "\n")[1:]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Errorf("\noutput mismatch (-have, +want)\n%s", diff)
				}
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	testCases := []struct {
		name     string
		cmdline  string
		wantCode int
	}{
		{
			name:     "success",
			cmdline:  "help",
			wantCode: 0,
		},
		{
			name:     "failure",
			cmdline:  "direct nonexisting",
			wantCode: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var out bytes.Buffer
			have := mainInt(&out, strings.Fields(tc.cmdline))
			want := tc.wantCode
			if have != want {
				t.Errorf("\nstatus code: have: %d; want: %d", have, want)
			}
		})
	}
}
