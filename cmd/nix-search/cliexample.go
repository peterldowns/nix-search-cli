package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// CLIExample is a helper for generating CLI example docs for cobra commands.  It
// removes any surrounding space from a string, then removes any leading
// whitespace from each line in the string. Any comments in the string will be
// colored as fainter than the rest of the text.
func CLIExample(s string) string {
	return example(s, "  ")
}

func CLIShort(s string) string {
	return example(s, "")
}

var DocsLink = color.New(color.Faint).Sprint( //nolint:gochecknoglobals
	"Docs: https://github.com/peterldowns/nix-search-cli",
)

func CLIHelp(s string) string {
	return DocsLink + "\n\nHelp:\n" + example(s, "  ")
}

func example(s, indent string) string {
	in := strings.Split(strings.TrimSpace(s), "\n")
	var out []string

	for _, x := range in {
		// x = strings.TrimSpace(x)
		parts := strings.SplitN(x, "#", 2)
		switch len(parts) {
		case 1:
			out = append(out, fmt.Sprintf("%s%s", indent, x))
		case 2:
			cmd := parts[0]
			cmt := parts[1]
			out = append(out, fmt.Sprintf(
				"%s%s%s",
				indent,
				cmd,
				color.New(color.Faint).Sprintf("#%s", cmt)),
			)
		default:
		}
	}
	return strings.Join(out, "\n")
}
