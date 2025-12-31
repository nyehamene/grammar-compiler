package main

import (
	"grammar/cmd"
	"os"
)

func main() {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr
	if code := cmd.Run(os.Args[1:], stdin, stdout, stderr); code != 0 {
		os.Exit(code)
	}
}
