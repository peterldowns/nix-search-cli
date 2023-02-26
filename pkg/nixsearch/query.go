package nixsearch

import (
	"encoding/json"
)

type Query struct {
	// Meta
	MaxResults int    // How many results, max, should be returned
	Channel    string // which nix-channel index to look at. mutually exclusive with Flakes.
	Flakes     bool   // if true, uses the flakes index instead. mutually exclusive with Channel.

	// Every query can combine multiple different matchers. At least one of
	// these fields must not be empty for the query to be processed.

	// Search matches the same way that search.nixos.org does.
	Search *MatchSearch
	// Program filters by binaries installed by the package.
	Program *MatchProgram
	// Name filters by the attribute name of the package.
	Name *MatchName
	// Version filters by the version of the package.
	Version *MatchVersion
	// QueryString filters by a custom ElasticSearch QueryString-syntax query.
	QueryString *MatchAdvanced
}

// IsEmpty returns false if any match has been set
func (query Query) IsEmpty() bool {
	if query.Search != nil {
		return false
	}
	if query.Program != nil {
		return false
	}
	if query.Name != nil {
		return false
	}
	if query.QueryString != nil {
		return false
	}
	if query.Version != nil {
		return false
	}
	return true
}

// Dict is a convenience helper for constructing JSON queries to send to Elasticsearch.
type Dict map[string]interface{}

// Payload returns a map[string]any that is ready to be serialized to JSON
// and sent to ElasticSearch.
func (query Query) Payload() ([]byte, error) {
	var must []any
	if query.Search != nil {
		must = append(must, query.Search)
	}
	if query.Name != nil {
		must = append(must, query.Name)
	}
	if query.Program != nil {
		must = append(must, query.Program)
	}
	if query.Version != nil {
		must = append(must, query.Version)
	}
	if query.QueryString != nil {
		must = append(must, query.QueryString)
	}
	return json.Marshal(Dict{
		"from": 0,
		"size": query.MaxResults,
		"sort": []Dict{
			{
				"_score":            "desc",
				"package_attr_name": "desc",
				"package_pversion":  "desc",
			},
		},
		"query": Dict{
			"bool": Dict{
				"must": must,
			},
		},
	})
}
