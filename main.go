package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/janderland/fdbq/app"
)

var flags app.Flags

func init() {
	flag.BoolVar(&flags.Write, "write", false, "allow write queries")
	flag.Parse()
}

func main() {
	if err := app.Run(flags, flag.Args()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
