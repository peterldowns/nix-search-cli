package nixsearch

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MatchSearch struct {
	Search string
}

func (m MatchSearch) MarshalJSON() ([]byte, error) {
	multiMatchName := "multi_match_" + strings.ReplaceAll(m.Search, " ", "_")
	queries := []Dict{
		{
			"multi_match": Dict{
				"type":  "cross_fields",
				"_name": multiMatchName,
				"query": m.Search,
				"fields": []string{
					"package_attr_name^9",
					"package_attr_name.*^5.3999999999999995",
					"package_programs^9",
					"package_programs.*^5.3999999999999995",
					"package_pname^6",
					"package_pname.*^3.5999999999999996",
					"package_description^1.3",
					"package_description.*^0.78",
					"package_pversion^1.3",
					"package_pversion.*^0.78",
					"package_longDescription^1",
					"package_longDescription.*^0.6",
					"flake_name^0.5",
					"flake_name.*^0.3",
					"flake_resolved.*^99",
				},
			},
		},
	}
	for _, term := range strings.Split(m.Search, " ") {
		queries = append(queries, Dict{
			"wildcard": Dict{
				"package_attr_name": Dict{
					"value":            fmt.Sprintf("*%s*", term),
					"case_insensitive": true,
				},
			},
		})
	}
	return json.Marshal(Dict{
		"dis_max": Dict{
			"tie_breaker": 0.7,
			"queries":     queries,
		},
	})
}

type MatchName struct {
	Name string
}

func (m MatchName) MarshalJSON() ([]byte, error) {
	return json.Marshal(Dict{
		"dis_max": Dict{
			"tie_breaker": 0.7,
			"queries": []Dict{
				{
					"wildcard": Dict{
						"package_attr_name": Dict{
							"value": m.Name + "*",
						},
					},
				},
				{
					"match": Dict{
						"package_programs": m.Name,
					},
				},
			},
		},
	})
}

type MatchProgram struct {
	Program string
}

func (m MatchProgram) MarshalJSON() ([]byte, error) {
	return json.Marshal(Dict{
		"dis_max": Dict{
			"tie_breaker": 0.7,
			"queries": []Dict{
				{
					"wildcard": Dict{
						"package_programs": Dict{
							"value": m.Program + "*",
						},
					},
				},
				{
					"match": Dict{
						"package_programs": m.Program,
					},
				},
			},
		},
	})
}

type MatchVersion struct {
	Version string
}

func (m MatchVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(Dict{
		"dis_max": Dict{
			"tie_breaker": 0.7,
			"queries": []Dict{
				{
					"wildcard": Dict{
						"package_pversion": Dict{
							"value": m.Version + "*",
						},
					},
				},
				{
					"match": Dict{
						"package_pversion": m.Version,
					},
				},
			},
		},
	})
}

type MatchQueryString struct {
	QueryString string
}

func (m MatchQueryString) MarshalJSON() ([]byte, error) {
	return json.Marshal(Dict{
		"query_string": Dict{
			"query": m.QueryString,
		},
	})
}
