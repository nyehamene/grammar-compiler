package cmd

import (
	"flag"
	"fmt"
	"grammar/check"
	"grammar/command"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func CheckCommand(args []string, stdout, stderr io.Writer) int {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	var help, fromStdin bool
	checkCmd.SetOutput(stderr)
	checkCmd.BoolVar(&help, "h", false, "Print this message.")
	checkCmd.BoolVar(&help, "help", false, "Print this message.")
	checkCmd.BoolVar(&fromStdin, "stdin", false, "Check code from stdin.")
	checkCmd.Parse(args)

	if help {
		fmt.Fprint(stdout, command.CheckUsage)
		return 0
	}

	if !fromStdin && checkCmd.NArg() == 0 {
		fmt.Fprint(stdout, command.CheckUsage)
		return 0
	}

	checker := check.NewChecker()
	var finalErr error

	if fromStdin {
		content, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading from stdin: %s\n", err)
			return 1
		}
		finalErr = checker.CheckSource(content, "<stdin>")
	} else {
		for _, path := range checkCmd.Args() {
			info, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(stderr, "Error accessing path %s: %s\n", path, err)
				return 1
			}

			if info.IsDir() {
				err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".grammar") {
						finalErr = checker.Check(filePath)
					}
					return nil
				})
				if err != nil {
					fmt.Fprintf(stderr, "Error walking directory %s: %s\n", path, err)
					return 1
				}
			} else {
				finalErr = checker.Check(path)
			}
		}
	}

	if finalErr != nil {
		if errs, ok := finalErr.(check.ErrorList); ok {
			errs.Format(stderr, checker.Sources())
		} else {
			fmt.Fprintln(stderr, finalErr)
		}
		return 1
	}

	return 0
}
