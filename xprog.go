// Package xprog contains helper functions for writing tests meant to be run via xprog.
//
// See the README for more information.
package xprog

import (
	"os"
)

// Absent returns true if the caller (the test) is not running from xprog.
// This is a safety measure that all destructive tests should follow.
// Meant to be called by tests as follows:
//
// func TestDemoXprog(t *testing.T) {
//     if xprog.Absent() {
//         t.Skip("skip: test requires xprog")
//     }
//     ...
// }
//
// See the README for more information.
func Absent() bool {
	return Target() == ""
}

// Target returns the xprog target URL.
func Target() string {
	return os.Getenv("XPROG_SYS_TARGET")
}
