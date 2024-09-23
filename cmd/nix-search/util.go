package main

import (
	"os"
	"strings"
)

// firstOf returns the first non-empty string from a slice of strings, stripped
// of all whitespace.
func firstOf(s ...string) string {
	for _, x := range s {
		x = strings.TrimSpace(x)
		if x != "" {
			return x
		}
	}
	return ""
}

// imap applies the function f to every element in the slice xs
func imap[X any, Y any](f func(x X) Y, xs []X) []Y {
	ys := make([]Y, 0, len(xs))
	for _, x := range xs {
		ys = append(ys, f(x))
	}
	return ys
}

// isTerminal will be true if we are outputting to a user shell. The value is
// set during init time to avoid unnecessary calls to Stat.
// The implementation is thanks to
// https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/
var isTerminal bool //nolint:gochecknoglobals
func init() { //nolint:gochecknoinits
	fileInfo, _ := os.Stdout.Stat()
	isTerminal = (fileInfo.Mode() & os.ModeCharDevice) == os.ModeCharDevice
}
