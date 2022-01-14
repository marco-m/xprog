package examples

import (
	"fmt"
	"os"
	"runtime"
)

// To be called by both tests: running on host and running on target via xprog.
func Harmless() {
	fmt.Fprintln(os.Stderr, "hello from Harmless on", runtime.GOOS)
}

// To be called only by tests running on target via xprog.
func Destructive() {
	fmt.Fprintln(os.Stderr, "hello from Destructive on", runtime.GOOS)
}
