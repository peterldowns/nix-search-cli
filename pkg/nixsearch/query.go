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
	QueryString *MatchQueryString
}

func (q Query) ExactlyMatches(program string) bool {
	if q.Program != nil && q.Program.Program == program {
		return true
	}
	if q.QueryString != nil && q.QueryString.QueryString == program {
		return true
	}
	if q.Name != nil && q.Name.Name == program {
		return true
	}
	if q.Search != nil && q.Search.Search == program {
		return true
	}
	return false
}

// IsEmpty returns false if any match has been set
func (q Query) IsEmpty() bool {
	if q.Search != nil {
		return false
	}
	if q.Program != nil {
		return false
	}
	if q.Name != nil {
		return false
	}
	if q.QueryString != nil {
		return false
	}
	if q.Version != nil {
		return false
	}
	return true
}

// Payload returns a map[string]any that is ready to be serialized to JSON
// and sent to ElasticSearch.
func (q Query) Payload() ([]byte, error) {
	must := []any{
		Dict{
			"match": Dict{
				"type": "package",
			},
		},
	}
	if q.Search != nil {
		must = append(must, q.Search)
	}
	if q.Name != nil {
		must = append(must, q.Name)
	}
	if q.Program != nil {
		must = append(must, q.Program)
	}
	if q.Version != nil {
		must = append(must, q.Version)
	}
	if q.QueryString != nil {
		must = append(must, q.QueryString)
	}
	return json.Marshal(Dict{
		"from": 0,
		"size": q.MaxResults,
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

// Dict is a convenience helper for constructing JSON queries to send to Elasticsearch.
type Dict map[string]interface{}
