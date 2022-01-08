# xprog

xprog -- a test runner for go test -exec.

Generic usage from go test:

    go test -exec='xprog <command> [opts] --' <go-packages> [go-test-flags]

Cross-compile the tests and run them on the target OS, connect via SSH:

    GOOS=linux go test -exec='xprog ssh [opts] --' <go-packages> [go-test-flags]

To see xprog output, pass -v both to xprog and go test:

    go test -v -exec='xprog -v <command> [opts] --' <go-packages> [go-test-flags]


## Limitations

`xprog ssh` expects a `ssh_config` file generated by `vagrant ssh-config` and will pick the first `Host` entry (see Examples below).


## Install

    go install github.com/marco-m/xprog@latest


## Examples

File [runner_test.go](runner_test.go) contains:

    func TestForXprog(t *testing.T) {
        t.Log("OS:", runtime.GOOS)
    }


### xprog direct

Use `xprog direct` to run the `TestForXprog` test on the host (in this case `xprog` is not even needed):

    $ go test -exec='xprog -v direct --' . -run TestForXprog -v
        runner_test.go:9: OS: darwin

### xprog ssh

Start the VM:

    virtualbox up

Generate the ssh_config file:

    vagrant ssh-config > ssh_config.vagrant

Cross-compile the tests and run them on the target OS, using `xprog ssh`:

    $ GOOS=linux go test -exec='xprog ssh --sshconfig ssh_config.vagrant --' . -run TestForXprog -v
        runner_test.go:9: OS: linux


## License

See [LICENSE](LICENSE).


## Credits

- [vmtest](https://github.com/anatol/vmtest)
- [dockexec](https://github.com/mvdan/dockexec)
- [Go tooling essentials](https://rakyll.org/go-tool-flags/)
- [go_android_exec](https://github.com/golang/go/blob/master/misc/android/go_android_exec.go)