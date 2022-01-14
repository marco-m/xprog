// Package xprog contains helper functions for writing tests meant to be run via xprog.
// Note: if a test package doesn't call one of the functions in this package, it will
// still have to import it with
//     import _ "github.com/marco-m/xprog"
// This is unfortunate but needed for the CLI flag machinery to work.
// See the README for why a less intrusive environment variable would not work.
package xprog

import "flag"

// xprog will invoke the test binary passing this flag, -xprog.target=...
var xprogTarget = flag.String("xprog.target", "", "")

// Target returns the xprog target URL. Meant to be called by tests as follows:
//
// func TestDemoXprog(t *testing.T) {
//     if xprog.Target() == "" {
//         t.Skip("skip: test requires xprog")
//     }
//     ...
// }
func Target() string {
	return *xprogTarget
	// return os.Getenv("XPROG_TARGET")
}
