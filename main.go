package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagSet = flag.NewFlagSet("xprog", flag.ContinueOnError)
	verbose = flagSet.Bool("verbose", false, "verbose output")
)

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage:

    go test -exec='xprog [docker run flags] image:tag' [test flags]

You can also run it directly, if you must:

    xprog image:tag [docker flags] pkg.test [test flags]

`)
	flagSet.PrintDefaults()
}

type usageErr string

func (ue usageErr) Error() string {
	return string(ue)
}

type flagErr string

func (fe flagErr) Error() string {
	return string(fe)
}

func main() {
	os.Exit(main2())
}

func main2() int {
	flagSet.Usage = usage
	err := main3()

	// runtime: no error
	if err == nil {
		return 0
	}

	// CLI parsing errors
	switch err.(type) {
	case usageErr:
		fmt.Fprintln(os.Stderr, err)
		flagSet.Usage()
		return 2
	case flagErr:
		return 2
	}

	// runtime: error
	fmt.Fprintln(os.Stderr, err)
	return 1
}

func main3() error {
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return flagErr(err.Error())
	}
	// args := flagSet.Args()

	if *verbose {
		fmt.Fprintln(os.Stderr, "args:")
		for _, a := range os.Args {
			fmt.Fprintln(os.Stderr, "  ", a)
		}
	}

	return nil
}
