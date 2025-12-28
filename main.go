package main

import (
	"grammar/cmd"
	"os"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:], os.Stdout, os.Stderr))
}
