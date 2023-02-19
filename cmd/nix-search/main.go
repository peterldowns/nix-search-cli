//nolint:gochecknoglobals
package main

import (
	"fmt"
	"strings"

	escapes "github.com/snugfox/ansi-escapes"
	"github.com/spf13/cobra"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

var root = &cobra.Command{
	Use:              "nix-search",
	Short:            "search for derivations via search.nixos.org",
	TraverseChildren: true,
	Args:             cobra.ArbitraryArgs,
	Run:              rootImpl,
}

var rootFlags struct {
	Query   *string
	Channel *string
}

func rootImpl(c *cobra.Command, args []string) {
	query := *rootFlags.Query
	if len(args) != 0 {
		if query != "" {
			fmt.Printf("[warning]: arbitrary arguments are being ignored due to explicit --query\n")
		} else {
			query = strings.Join(args, " ")
		}
	}
	if query == "" {
		_ = c.Usage()
		return
	}

	channel := *rootFlags.Channel
	results, err := nixsearch.Search(nixsearch.Input{
		Channel: channel,
		Query:   query,
	})
	if err != nil {
		panic(fmt.Errorf("failed search: %w", err))
	}
	for _, pkg := range results.Packages {
		url := fmt.Sprintf(`https://search.nixos.org/packages?channel=%s&show=%s`, results.Input.Channel, pkg.AttrName)
		fmt.Printf("%s", escapes.Link(url, pkg.AttrName))
		if len(pkg.Programs) != 0 {
			fmt.Printf(" -> [%s]", strings.Join(pkg.Programs, ", "))
		}
		fmt.Printf("\n")
	}
}

func main() {
	// Disable the builtin shell-completion script generator command
	root.CompletionOptions.DisableDefaultCmd = true
	rootFlags.Query = root.Flags().String("query", "", "the text to search for")
	rootFlags.Channel = root.Flags().String("channel", "unstable", "which channel to search in")

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
