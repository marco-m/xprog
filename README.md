# xprog

xprog -- a test runner for "go test -exec"

usage with go test:

    go test -exec='xprog <xprog-args>' <packages> <flags>

to see xprog output, pass -v both to xprog and go test:

    go test -v -exec='xprog -v <xprog-args>' <packages> <flags>


...


## Testing

    ln -s ~/src/xprog/bin/xprog ~/bin

    go build -o ./bin/xprog . && go test -v -exec='xprog <xprog-args>' .


## Credits

- [vmtest](https://github.com/anatol/vmtest)
- [dockexec](https://github.com/mvdan/dockexec)
- [Go tooling essentials](https://rakyll.org/go-tool-flags/)
- [go_android_exec](https://github.com/golang/go/blob/master/misc/android/go_android_exec.go)