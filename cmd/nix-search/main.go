//nolint:gochecknoglobals
package main

import (
	"fmt"
	"strings"

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

func rootImpl(_ *cobra.Command, args []string) {
	var query string
	if len(args) != 0 {
		query = strings.Join(args, " ")
		if *rootFlags.Query != "" {
			fmt.Printf("[warning]: arbitrary arguments are being overrideen by --query\n")
			query = *rootFlags.Query
		}
	}

	channel := *rootFlags.Channel
	fmt.Printf("query = %s\n", query)
	fmt.Printf("channel = %s\n", channel)
	results, err := nixsearch.Search(nixsearch.Input{
		Channel: channel,
		Query:   query,
	})
	if err != nil {
		panic(fmt.Errorf("failed search: %w\n", err))
	}
	fmt.Printf("Found %d results\n", len(results.Derivations))
	for _, derivation := range results.Derivations {
		fmt.Println(derivation)
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
