package check

import "log"

type Option func(*Checker)

func SetLogger(l *log.Logger) func(*Checker) {
	return func(c *Checker) {
		logger := log.New(l.Writer(), "check: ", l.Flags())
		c.log = logger
	}
}

func SetFileLoader(l FileLoader) func(*Checker) {
	return func(c *Checker) {
		c.cu.loader = l
	}
}
