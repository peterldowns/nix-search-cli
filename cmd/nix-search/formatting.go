package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	escapes "github.com/snugfox/ansi-escapes"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

func formatPackageName(isTerminal bool, input nixsearch.Query, pkg nixsearch.Package) string {
	c := color.New(color.Underline, color.FgBlue)
	if input.Flakes {
		var name string
		switch pkg.FlakeResolved.Type {
		case "github":
			name = c.Sprintf(
				"%s:%s/%s#%s",
				pkg.FlakeResolved.Type,
				pkg.FlakeResolved.Owner,
				pkg.FlakeResolved.Repo,
				pkg.AttrName,
			)
		case "git":
			name = c.Sprintf("%s#%s", pkg.FlakeResolved.URL, pkg.AttrName)
		default:
			name = "unknown:" + pkg.FlakeName
		}
		if isTerminal {
			url := fmt.Sprintf(`https://search.nixos.org/flakes?show=%s&query=%s`, pkg.AttrName, pkg.AttrName)
			return escapes.Link(url, name)
		}
		return name
	}
	if isTerminal {
		url := fmt.Sprintf(`https://search.nixos.org/packages?channel=%s&show=%s`, input.Channel, pkg.AttrName)
		return escapes.Link(url, c.Sprint(pkg.AttrName))
	}
	return pkg.AttrName
}

func formatDependencies(isTerminal bool, input nixsearch.Query, programs []string) string {
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

func isMatch(input nixsearch.Query, program string) bool {
	return ((input.Program != nil && input.Program.Program == program) ||
		(input.QueryString != nil && input.QueryString.Advanced == program) ||
		(input.Name != nil && input.Name.Name == program) ||
		(input.Search != nil && input.Search.Search == program))
}

func printResults(input nixsearch.Query, packages []nixsearch.Package) {
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
		name := formatPackageName(isTerminal, input, pkg)
		fmt.Print(name)
		if !showDetails {
			vstring := color.New(color.FgGreen, color.Faint).Sprintf("@ %s", pkg.Version)
			fmt.Print(" ", vstring)
			if len(pkg.Programs) != 0 {
				programs := formatDependencies(isTerminal, input, pkg.Programs)
				fmt.Print(": ", programs)
			}
			fmt.Println()
			continue
		}
		fmt.Println()
		// version
		fmt.Printf("  version: %s\n", pkg.Version)
		// programs
		fmt.Printf("  programs: %s\n", formatDependencies(isTerminal, input, pkg.Programs))
		// description
		d := firstOf(pkg.FlakeDescription, pkg.Description)
		fmt.Printf("  description: %s\n", d)
		// license
		fmt.Printf("  license:")
		if len(pkg.Licenses) == 1 {
			license := pkg.Licenses[0]
			txt := license.FullName
			if isTerminal && license.URL != "" {
				txt = escapes.Link(license.URL, license.FullName)
			}
			fmt.Printf(" %s\n", txt)
		} else {
			fmt.Printf("\n")
			for _, license := range pkg.Licenses {
				txt := license.FullName
				if isTerminal && license.URL != "" {
					txt = escapes.Link(license.URL, license.FullName)
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
