//nolint:gochecknoglobals
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	Query   *string
	Channel *string
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

	// If the user doesn't pass --query and they don't pass any positional
	// arguments, show the usage and exit since there is no defined search term.
	if query == "" {
		_ = c.Usage()
		return
	}

	ctx := context.Background()
	client, err := nixsearch.NewClient()
	if err != nil {
		panic(fmt.Errorf("failed to load search client: %w", err))
	}

	packages, err := client.Search(ctx, channel, query)
	if err != nil {
		panic(fmt.Errorf("failed search: %w", err))
	}

	// thanks https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/
	o, _ := os.Stdout.Stat()
	isTerminal := (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice

	for _, pkg := range packages {
		if isTerminal {
			url := fmt.Sprintf(`https://search.nixos.org/packages?channel=%s&show=%s`, channel, pkg.AttrName)
			fmt.Printf("%s", escapes.Link(url, pkg.AttrName))
		} else {
			fmt.Printf("%s", pkg.AttrName)
		}
		if len(pkg.Programs) != 0 {
			fmt.Printf(" -> [%s]", strings.Join(pkg.Programs, ", "))
		}
		fmt.Printf("\n")
	}
}

func main() {
	// Disable the builtin shell-completion script generator command
	rootCommand.CompletionOptions.DisableDefaultCmd = true
	rootFlags.Query = rootCommand.Flags().String("query", "", "the text to search for")
	rootFlags.Channel = rootCommand.Flags().String("channel", "unstable", "which channel to search in")

	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}
