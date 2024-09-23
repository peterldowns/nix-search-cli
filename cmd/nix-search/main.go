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
	Use:   "nix-search some program or package [flags]",
	Short: DocsLink,
	Example: CLIExample(`
# Search for nix packages in the https://search.nixos.org index

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
# ... on a specific channel, default "unstable". The valid channel
#     values are what the search.nixos.org index has, check
#     that website to see what options they show in their interface.
nix-search --channel=unstable python3
# ... or flakes indexed by search.nixos.org, see their website
#     for more information.
nix-search --flakes wayland

# ... or search with multiple filters and options
nix-search golang --program go --version '1.*' --details
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
	Reverse     *bool
}

func root(c *cobra.Command, args []string) error {
	channel := *rootFlags.Channel
	search := *rootFlags.Search
	if len(args) != 0 && search == "" {
		search = strings.Join(args, " ")
	}

	query := nixsearch.Query{
		Channel:    channel,
		Flakes:     *rootFlags.Flakes,
		MaxResults: *rootFlags.MaxResults,
	}
	if x := search; x != "" {
		query.Search = &nixsearch.MatchSearch{Search: x}
	}
	if x := *rootFlags.Program; x != "" {
		query.Program = &nixsearch.MatchProgram{Program: x}
	}
	if x := *rootFlags.Name; x != "" {
		query.Name = &nixsearch.MatchName{Name: x}
	}
	if x := *rootFlags.Version; x != "" {
		query.Version = &nixsearch.MatchVersion{Version: x}
	}
	if x := *rootFlags.QueryString; x != "" {
		query.QueryString = &nixsearch.MatchQueryString{QueryString: x}
	}

	// If the user doesn't give any search terms or any flags, show the
	// program's usage information and exit.
	if query.IsEmpty() {
		return c.Help()
	}

	ctx := context.Background()
	client, err := nixsearch.NewElasticSearchClient()
	if err != nil {
		return err
	}

	packages, err := client.Search(ctx, query)
	if err != nil {
		return err
	}

	printResults(query, packages)
	return nil
}

func main() {
	rootCommand.CompletionOptions.DisableDefaultCmd = true // Disable the builtin shell-completion script generator command
	rootCommand.SilenceErrors = true
	rootCommand.SilenceUsage = true
	rootCommand.TraverseChildren = true

	rootFlags.Search = rootCommand.Flags().StringP("search", "s", "", "default search, same as the website")
	rootFlags.Channel = rootCommand.Flags().StringP("channel", "c", "unstable", "which channel to search in")
	rootFlags.Program = rootCommand.Flags().StringP("program", "p", "", "search by installed programs")
	rootFlags.Name = rootCommand.Flags().StringP("name", "n", "", "search by package name")
	rootFlags.QueryString = rootCommand.Flags().StringP("query-string", "q", "", "search by elasticsearch querystring")
	rootFlags.JSON = rootCommand.Flags().BoolP("json", "j", false, "emit results in json-line format")
	rootFlags.Details = rootCommand.Flags().BoolP("details", "d", false, "show expanded details for each result")
	rootFlags.MaxResults = rootCommand.Flags().IntP("max-results", "m", 20, "maximum number of results to return")
	rootFlags.Reverse = rootCommand.Flags().BoolP("reverse", "r", false, "print results in reverse order")
	rootFlags.Version = rootCommand.Flags().StringP("version", "v", "", "search by version")
	rootFlags.Flakes = rootCommand.Flags().BoolP("flakes", "f", false, "search flakes instead of nixpkgs")

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
