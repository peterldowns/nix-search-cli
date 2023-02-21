//nolint:gochecknoglobals
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	escapes "github.com/snugfox/ansi-escapes"
	"github.com/spf13/cobra"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

var rootCommand = &cobra.Command{
	Use:              "nix-search",
	Short:            "search for derivations via search.nixos.org",
	TraverseChildren: true,
	Args:             cobra.ArbitraryArgs,
	Run:              root,
}

var rootFlags struct {
	Channel     *string
	Search      *string
	Program     *string
	Attr        *string
	QueryString *string
	JSON        *bool
	Details     *bool
	MaxResults  *int
	Version     *string
}

func root(c *cobra.Command, args []string) {
	channel := *rootFlags.Channel
	query := *rootFlags.Search
	if len(args) != 0 {
		if query != "" {
			fmt.Printf("[warning]: arbitrary arguments are being ignored due to explicit --query\n")
		} else {
			query = strings.Join(args, " ")
		}
	}
	input := nixsearch.Input{
		Channel:    channel,
		Default:    query,
		Program:    *rootFlags.Program,
		Name:       *rootFlags.Attr,
		Advanced:   *rootFlags.QueryString,
		MaxResults: *rootFlags.MaxResults,
		Version:    *rootFlags.Version,
	}

	// If the user doesn't pass --query and they don't pass any positional
	// arguments, show the usage and exit since there is no defined search term.
	if input.Default == "" && input.Program == "" && input.Name == "" && input.Advanced == "" && input.Version == "" {
		_ = c.Usage()
		return
	}

	ctx := context.Background()
	client, err := nixsearch.NewClient()
	if err != nil {
		panic(fmt.Errorf("failed to load search client: %w", err))
	}

	packages, err := client.Search(ctx, input)
	if err != nil {
		panic(fmt.Errorf("failed search: %w", err))
	}

	// thanks https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/
	o, _ := os.Stdout.Stat()
	isTerminal := (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice

	showDetails := *rootFlags.Details

	for _, pkg := range packages {
		// If asked to spit out json, just dump the packages directly
		if rootFlags.JSON != nil && *rootFlags.JSON {
			line, _ := json.Marshal(pkg)
			fmt.Println(string(line))
			continue
		}
		name := formatPackageName(isTerminal, channel, pkg.AttrName)
		fmt.Print(name)
		if input.Version != "" {
			vstring := color.New(color.FgGreen, color.Faint).Sprint(pkg.Version)
			fmt.Print(" ", vstring)
		}
		if len(pkg.Programs) != 0 {
			programs := formatDependencies(isTerminal, input, pkg.Programs)
			fmt.Print(": ", programs)
		}
		fmt.Println()
		if !showDetails {
			continue
		}
		// version
		fmt.Printf("  version: %s\n", pkg.Version)
		// description
		d := ""
		if pkg.Description != nil {
			d = *pkg.Description
		}
		fmt.Printf("  description: %s\n", d)
		// license
		fmt.Printf("  license:")
		if len(pkg.Licenses) == 1 {
			license := pkg.Licenses[0]
			txt := license.FullName
			if isTerminal && license.URL != nil {
				txt = escapes.Link(*license.URL, license.FullName)
			}
			fmt.Printf(" %s\n", txt)
		} else {
			fmt.Printf("\n")
			for _, license := range pkg.Licenses {
				txt := license.FullName
				if isTerminal && license.URL != nil {
					txt = escapes.Link(*license.URL, license.FullName)
				}
				fmt.Printf("    - %s\n", txt)
			}
		}
		// homepage
		fmt.Printf("  homepage:")
		if len(pkg.Homepage) == 1 {
			fmt.Printf(" %s\n", pkg.Homepage[0])
		} else {
			fmt.Printf("\n")
			for _, homepage := range pkg.Homepage {
				fmt.Printf("    - %s\n", homepage)
			}
		}
	}
}

func formatPackageName(isTerminal bool, channel, attrName string) string {
	if isTerminal {
		c := color.New(color.Underline, color.FgBlue)
		url := fmt.Sprintf(`https://search.nixos.org/packages?channel=%s&show=%s`, channel, attrName)
		return escapes.Link(url, c.Sprint(attrName))
	}
	return attrName
}

func formatDependencies(isTerminal bool, input nixsearch.Input, programs []string) string {
	if isTerminal {
		var matches []string
		var others []string
		// Dim all the programs that aren't what you searched for
		for _, program := range programs {
			if isMatch(input, program) {
				matches = append(matches, color.New(color.Bold).Sprint(program))
			} else {
				others = append(others, color.New(color.Faint).Sprint(program))
			}
		}
		sort.Strings(matches)
		sort.Strings(others)
		matches = append(matches, others...)
		programs = matches
	}
	return strings.Join(programs, " ")
}

func isMatch(input nixsearch.Input, program string) bool {
	return (input.Program == program ||
		input.Advanced == program ||
		input.Name == program ||
		input.Default == program)
}

func main() {
	// Disable the builtin shell-completion script generator command
	rootCommand.CompletionOptions.DisableDefaultCmd = true
	rootFlags.Search = rootCommand.Flags().StringP("search", "s", "", "default search, same as the website")
	rootFlags.Channel = rootCommand.Flags().StringP("channel", "c", "unstable", "which channel to search in")
	rootFlags.Program = rootCommand.Flags().StringP("program", "p", "", "search by installed programs")
	rootFlags.Attr = rootCommand.Flags().StringP("attr", "a", "", "search by attr name")
	rootFlags.QueryString = rootCommand.Flags().StringP("query-string", "q", "", "perform an advanced query string format search")
	rootFlags.JSON = rootCommand.Flags().BoolP("json", "j", false, "emit results in json-line format")
	rootFlags.Details = rootCommand.Flags().BoolP("details", "d", false, "show expanded details for each result")
	rootFlags.MaxResults = rootCommand.Flags().IntP("max-results", "m", 20, "maximum number of results to return")
	rootFlags.Version = rootCommand.Flags().StringP("version", "v", "", "search by version")

	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}
