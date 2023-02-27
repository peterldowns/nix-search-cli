//nolint:gochecknoglobals
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

var rootCommand = &cobra.Command{
	Use:   "nix-search ...query",
	Short: "search for packages via search.nixos.org",
	Example: trimLeading(`
# Search
# ... like the web interface
nix-search python linter
nix-search --search "python linter"  
# ... by package name
nix-search --name python
nix-search --name 'emacsPackages.*'  
# ... by version
nix-search --version 1.20 
nix-search --version '1.*'           
# ... by installed programs
nix-search --program python
nix-search --program "py*"
# ... with ElasticSearch QueryString syntax
nix-search --query-string="package_programs:(crystal OR irb)"
nix-search --query-string='package_description:(MIT Scheme)'
# ... on a specific channel
nix-search --channel=unstable python3
# ... or flakes
nix-search --flakes wayland
# ... with multiple filters and options
nix-search --name go --version 1.20 --details
	`),
	TraverseChildren: true,
	Args:             cobra.ArbitraryArgs,
	RunE:             root,
}

var rootFlags struct {
	Channel     *string
	Flakes      *bool
	Search      *string
	Program     *string
	Name        *string
	Version     *string
	QueryString *string
	JSON        *bool
	Details     *bool
	MaxResults  *int
}

func root(c *cobra.Command, args []string) error {
	channel := *rootFlags.Channel
	query := *rootFlags.Search
	if len(args) != 0 {
		if query != "" {
			fmt.Printf("[warning]: arbitrary arguments are being ignored due to explicit --query\n")
		} else {
			query = strings.Join(args, " ")
		}
	}

	input := nixsearch.Query{
		Channel:    channel,
		Flakes:     *rootFlags.Flakes,
		MaxResults: *rootFlags.MaxResults,
	}
	if x := query; x != "" {
		input.Search = &nixsearch.MatchSearch{Search: x}
	}
	if x := *rootFlags.Program; x != "" {
		input.Program = &nixsearch.MatchProgram{Program: x}
	}
	if x := *rootFlags.Name; x != "" {
		input.Name = &nixsearch.MatchName{Name: x}
	}
	if x := *rootFlags.Version; x != "" {
		input.Version = &nixsearch.MatchVersion{Version: x}
	}
	if x := *rootFlags.QueryString; x != "" {
		input.QueryString = &nixsearch.MatchQueryString{QueryString: x}
	}

	// If the user doesn't give any search terms or any flags, show the
	// program's usage information and exit.
	if input.IsEmpty() {
		_ = c.Help()
		return nil
	}

	ctx := context.Background()
	client, err := nixsearch.NewElasticSearchClient()
	if err != nil {
		return err
	}

	packages, err := client.Search(ctx, input)
	if err != nil {
		return err
	}

	printResults(input, packages)
	return nil
}

func main() {
	rootCommand.CompletionOptions.DisableDefaultCmd = true // Disable the builtin shell-completion script generator command
	rootFlags.Search = rootCommand.Flags().StringP("search", "s", "", "default search, same as the website")
	rootFlags.Channel = rootCommand.Flags().StringP("channel", "c", "unstable", "which channel to search in")
	rootFlags.Program = rootCommand.Flags().StringP("program", "p", "", "search by installed programs")
	rootFlags.Name = rootCommand.Flags().StringP("name", "n", "", "search by package name")
	rootFlags.QueryString = rootCommand.Flags().StringP("query-string", "q", "", "search by elasticsearch querystring")
	rootFlags.JSON = rootCommand.Flags().BoolP("json", "j", false, "emit results in json-line format")
	rootFlags.Details = rootCommand.Flags().BoolP("details", "d", false, "show expanded details for each result")
	rootFlags.MaxResults = rootCommand.Flags().IntP("max-results", "m", 20, "maximum number of results to return")
	rootFlags.Version = rootCommand.Flags().StringP("version", "v", "", "search by version")
	rootFlags.Flakes = rootCommand.Flags().BoolP("flakes", "f", false, "search flakes instead of nixpkgs")
	rootCommand.SilenceErrors = true
	rootCommand.SilenceUsage = true

	defer func() {
		switch r := recover().(type) {
		case error:
			onError(r)
		default:
		}
	}()

	if err := rootCommand.Execute(); err != nil {
		onError(err)
	}
}

func onError(err error) {
	errstr := color.New(color.FgRed, color.Italic).Sprint("error: ", err.Error())
	fmt.Fprintln(os.Stderr, "\n", errstr)
	os.Exit(1)
}
