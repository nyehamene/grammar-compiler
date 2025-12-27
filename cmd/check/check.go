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
	var hadError bool

	handleCheckError := func(err error) {
		if err != nil {
			// This will be expanded later to handle ast.ErrorList
			fmt.Fprintln(stderr, err)
			hadError = true
		}
	}

	if fromStdin {
		content, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading from stdin: %s\n", err)
			return 1
		}
		handleCheckError(checker.CheckSource(content, "<stdin>"))
	} else {
		for _, path := range checkCmd.Args() {
			info, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(stderr, "Error accessing path %s: %s\n", path, err)
				hadError = true
				continue
			}

			if info.IsDir() {
				err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
					if err != nil {
						return err // Propagate errors from walking the path.
					}
					if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".grammar") {
						handleCheckError(checker.Check(filePath))
					}
					return nil
				})
				if err != nil {
					fmt.Fprintf(stderr, "Error walking directory %s: %s\n", path, err)
					hadError = true
				}
			} else {
				handleCheckError(checker.Check(path))
			}
		}
	}

	if hadError {
		return 1
	}

	// We can add a success message if needed, for now, silence is success.
	return 0
}
