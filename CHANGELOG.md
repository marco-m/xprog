# xprog Changelog

# [v0.3.0] - Unreleased

## New

- ssh: add flag `--sudo` to run the test binary with sudo.
- Function `xprog.Absent`: support skipping destructive tests on the host.
  See explanation in README and examples in directory `examples` (supersedes function `xprog.Target`).
- Flag `--version` prints the xprog version.

# [v0.2.0] - [2022-01-14]

## Breaking changes

- No outside impact: shuffled packages around to make it possible for client code to import the top-level module name (`import "github.com/marco-m/xprog"`).

## Changes

- Better documentation.

## New

- Function `xprog.Target`: support skipping destructive tests on the host.
  See explanation in README and examples in directory `examples` (superseded by function `xprog.Absent` in v0.3.0).

- Add examples in directory `examples`, see also walk-through in README.

# [v0.1.0] - [2022-01-10]

## New

- Support collecting code coverage when invoked from `go test -coverprofile`. This means that one can then call `go tool cover -html` on the host :-)

# [v0.0.1] - [2022-01-06]

## New

- First release. Support `direct` and `ssh` subcommands.


[v0.0.1]: https://github.com/marco-m/xprog/releases/tag/v0.0.1
[v0.1.0]: https://github.com/marco-m/xprog/releases/tag/v0.1.0
[v0.2.0]: https://github.com/marco-m/xprog/releases/tag/v0.2.0
[v0.3.0]: https://github.com/marco-m/xprog/releases/tag/v0.3.0
