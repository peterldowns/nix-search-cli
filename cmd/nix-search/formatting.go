package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	escapes "github.com/snugfox/ansi-escapes"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

func printResults(input nixsearch.Query, packages []nixsearch.Package) {
	for _, pkg := range packages {
		printResult(input, pkg)
	}
}

func printResult(input nixsearch.Query, pkg nixsearch.Package) {
	// { ... pkg contents ... }
	shouldOutputJSON := (rootFlags.JSON != nil && *rootFlags.JSON)
	if shouldOutputJSON {
		bytes, _ := json.Marshal(pkg)
		fmt.Println(string(bytes))
		return
	}

	// name @ version: program1 program2 ...
	shouldOutputDetails := (rootFlags.Details != nil && *rootFlags.Details)
	if !shouldOutputDetails {
		name := formatName(input, pkg)
		fmt.Print(name)
		if pkg.Version != "" {
			version := formatVersion("@ " + pkg.Version)
			fmt.Print(" ", version)
		}
		if len(pkg.Programs) > 0 {
			programs := formatPrograms(input, pkg.Programs)
			fmt.Print(" : ", programs)
		}
		fmt.Print("\n")
		return
	}

	// examplePkg
	//  version: 3.1
	//  programs: hello goodbye true false
	//  description: examplePkg is a made up package as an example.
	//  homepage: [https://example.com]
	//    - https://example.com/one
	//    - https://example.org/two
	//  license: [Single License Result Shown On One Line]
	//    - Multiple License
	//    - Results Shown Over Multiple Lines
	name := formatName(input, pkg)
	fmt.Print(name, "\n")
	version := formatVersion(pkg.Version)
	fmt.Printf("  version: %s\n", version)
	programs := formatPrograms(input, pkg.Programs)
	fmt.Printf("  programs: %s\n", programs)
	description := formatDescription(pkg)
	fmt.Printf("  description: %s\n", description)
	fmt.Print("  homepage:")
	fmt.Print(formatList(imap(func(s string) string {
		return formatLink(s, s, color.Underline)
	}, pkg.Homepage)))
	fmt.Print("  license:")
	fmt.Print(formatList(imap(formatLicense, pkg.Licenses)))
}

func formatName(input nixsearch.Query, pkg nixsearch.Package) string {
	if pkg.IsFlake() {
		var name string
		switch pkg.FlakeResolved.Type {
		case "github":
			name = fmt.Sprintf(
				"%s:%s/%s#%s",
				pkg.FlakeResolved.Type,
				pkg.FlakeResolved.Owner,
				pkg.FlakeResolved.Repo,
				pkg.AttrName,
			)
		case "git":
			name = fmt.Sprintf("%s#%s", pkg.FlakeResolved.URL, pkg.AttrName)
		default:
			name = "unknown:" + pkg.FlakeName
		}
		url := fmt.Sprintf(
			`https://search.nixos.org/flakes?show=%s&query=%s`,
			pkg.AttrName,
			pkg.AttrName,
		)
		return formatLink(url, name, color.Underline, color.FgWhite)
	}
	url := fmt.Sprintf(`https://search.nixos.org/packages?channel=%s&show=%s`, input.Channel, pkg.AttrName)
	return formatLink(url, pkg.AttrName)
}

func formatVersion(version string) string {
	return color.New(color.FgGreen, color.Faint).Sprint(version)
}

func formatPrograms(input nixsearch.Query, programs []string) string {
	if len(programs) == 0 {
		return ""
	}
	if isTerminal {
		var matches []string
		var others []string
		// Dim all the programs that aren't what you searched for
		for _, program := range programs {
			if input.ExactlyMatches(program) {
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

func formatDescription(pkg nixsearch.Package) string {
	return firstOf(pkg.FlakeDescription, pkg.Description)
}

func formatLicense(license nixsearch.License) string {
	return formatLink(license.URL, license.FullName, color.Underline)
}

func formatLink(url, text string, attrs ...color.Attribute) string {
	if url != "" {
		var c *color.Color
		if attrs != nil {
			// Optional styling takes precedence
			c = color.New(attrs...)
		} else {
			// Default styling
			c = color.New(color.Underline, color.FgBlue)
		}
		if isTerminal {
			return escapes.Link(url, c.Sprint(text))
		}
		return c.Sprint(text)
	}
	return text
}

func formatList(data []string) string {
	if len(data) == 0 {
		return "\n"
	}
	if len(data) == 1 {
		return fmt.Sprintf(" %s\n", data[0])
	}
	fmt.Printf("\n")
	out := strings.Builder{}
	out.WriteString("\n")
	for _, s := range data {
		out.WriteString(fmt.Sprintf("    - %s\n", s))
	}
	return out.String()
}
