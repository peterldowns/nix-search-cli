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
	Channel  *string
	Query    *string
	Program  *string
	Name     *string
	Advanced *string
	JSON     *bool
}

func root(c *cobra.Command, args []string) {
	channel := *rootFlags.Channel
	query := *rootFlags.Query
	if len(args) != 0 {
		if query != "" {
			fmt.Printf("[warning]: arbitrary arguments are being ignored due to explicit --query\n")
		} else {
			query = strings.Join(args, " ")
		}
	}
	input := nixsearch.Input{
		Channel:  channel,
		Default:  query,
		Program:  *rootFlags.Program,
		Name:     *rootFlags.Name,
		Advanced: *rootFlags.Advanced,
	}

	// If the user doesn't pass --query and they don't pass any positional
	// arguments, show the usage and exit since there is no defined search term.
	if input.Default == "" && input.Program == "" && input.Name == "" && input.Advanced == "" {
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

	showVersion := true
	showDescription := true
	showLicenses := true
	showHomepage := true

	for _, pkg := range packages {
		// If asked to spit out json, just dump the packages directly
		if rootFlags.JSON != nil && *rootFlags.JSON {
			line, _ := json.Marshal(pkg)
			fmt.Println(string(line))
			continue
		}
		name := formatPackageName(isTerminal, channel, pkg.AttrName)
		fmt.Print(name)
		if len(pkg.Programs) != 0 {
			programs := formatDependencies(isTerminal, input, pkg.Programs)
			fmt.Print(": ", programs)
		}
		fmt.Println()
		if showVersion {
			fmt.Printf("  version: %s\n", pkg.Version)
		}
		if showDescription {
			d := ""
			if pkg.Description != nil {
				d = *pkg.Description
			}
			fmt.Printf("  description: %s\n", d)
		}
		if showLicenses {
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
		}
		if showHomepage {
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
	rootFlags.Query = rootCommand.Flags().StringP("query", "q", "", "default fuzzy search")
	rootFlags.Channel = rootCommand.Flags().StringP("channel", "c", "unstable", "which channel to search in")
	rootFlags.Program = rootCommand.Flags().StringP("program", "p", "", "search by installed programs")
	rootFlags.Name = rootCommand.Flags().StringP("name", "n", "", "search by attr name")
	rootFlags.Advanced = rootCommand.Flags().StringP("advanced", "a", "", "perform an advanced query string format search")
	rootFlags.JSON = rootCommand.Flags().BoolP("json", "j", false, "emit results in json-line format")

	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}
