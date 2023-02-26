package main

import (
	"strings"

	"github.com/fatih/color"
)

// trimLeading removes any surrounding space from a string, then removes any
// leading whitespace from each line in the string.
func trimLeading(s string) string {
	in := strings.Split(strings.TrimSpace(s), "\n")
	var out []string

	for _, x := range in {
		x = strings.TrimSpace(x)
		if len(x) > 0 && x[0] == '#' {
			x = color.New(color.Faint).Sprint(x)
		}
		out = append(out, "  "+x)
	}
	return strings.Join(out, "\n")
}

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
