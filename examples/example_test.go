package examples_test

import (
	"testing"

	"github.com/marco-m/xprog"
	"github.com/marco-m/xprog/examples"
)

// We want this test to run both on the host and on the target, so we simply don't add
// any conditional.
func TestHarmless(t *testing.T) {
	examples.Harmless()
}

// This shows how to run a test only on the target.
// All destructive or invasive tests (test with side-effects) should be protected in this
// way.
func TestDestructiveXprog(t *testing.T) {
	if xprog.Target() == "" {
		t.Skip("skip: test requires xprog")
	}
	examples.Destructive()
}
