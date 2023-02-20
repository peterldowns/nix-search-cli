package nixsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	HTTPClient *http.Client
}

func NewClient() (*Client, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = nil

	return &Client{
		HTTPClient: retryClient.StandardClient(),
	}, nil
}

func (c Client) Search(ctx context.Context, channel string, query string) ([]Package, error) {
	req, err := buildRequest(ctx, channel, query)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	packages, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}
	return packages, nil
}

func buildRequest(ctx context.Context, channel string, query string) (*http.Request, error) {
	url := formatURL(channel)
	payload, err := formatQuery(query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(defaultUsername, defaultPassword)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func parseResponse(resp *http.Response) ([]Package, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r Response
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, r.Error
	}

	packages := make([]Package, len(r.Hits.Hits))
	for i, hit := range r.Hits.Hits {
		packages[i] = hit.Package
	}
	return packages, nil
}

func formatURL(channel string) string {
	return fmt.Sprintf(urlTemplate, url.QueryEscape(channel))
}

func formatQuery(query string) (string, error) {
	matchName := "multi_match_" + strings.ReplaceAll(query, " ", "_")
	encQuery, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	encMatchName, err := json.Marshal(matchName)
	if err != nil {
		return "", err
	}
	value := fmt.Sprintf("*%s*", query)
	encValue, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(payloadTemplate, encQuery, encMatchName, encValue), nil
}

const (
	// https://github.com/NixOS/nixos-search/blob/main/frontend/src/index.js
	defaultUsername = "aWVSALXpZv"
	defaultPassword = "X8gPHnzL52wFEekuxsfQ9cSh"
	urlTemplate     = `https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net:443/latest-37-nixos-%s/_search`
	payloadTemplate = `
{
	"from": 0,
	"size": 50,
	"sort": [
	  {
		"_score": "desc",
		"package_attr_name": "desc",
		"package_pversion": "desc"
	  }
	],
	"aggs": {
	  "package_attr_set": {
		"terms": {
		  "field": "package_attr_set",
		  "size": 20
		}
	  },
	  "package_license_set": {
		"terms": {
		  "field": "package_license_set",
		  "size": 20
		}
	  },
	  "package_maintainers_set": {
		"terms": {
		  "field": "package_maintainers_set",
		  "size": 20
		}
	  },
	  "package_platforms": {
		"terms": {
		  "field": "package_platforms",
		  "size": 20
		}
	  },
	  "all": {
		"global": {},
		"aggregations": {
		  "package_attr_set": {
			"terms": {
			  "field": "package_attr_set",
			  "size": 20
			}
		  },
		  "package_license_set": {
			"terms": {
			  "field": "package_license_set",
			  "size": 20
			}
		  },
		  "package_maintainers_set": {
			"terms": {
			  "field": "package_maintainers_set",
			  "size": 20
			}
		  },
		  "package_platforms": {
			"terms": {
			  "field": "package_platforms",
			  "size": 20
			}
		  }
		}
	  }
	},
	"query": {
	  "bool": {
		"filter": [
		  {
			"term": {
			  "type": {
				"value": "package",
				"_name": "filter_packages"
			  }
			}
		  },
		  {
			"bool": {
			  "must": [
				{
				  "bool": {
					"should": []
				  }
				},
				{
				  "bool": {
					"should": []
				  }
				},
				{
				  "bool": {
					"should": []
				  }
				},
				{
				  "bool": {
					"should": []
				  }
				}
			  ]
			}
		  }
		],
		"must": [
		  {
			"dis_max": {
			  "tie_breaker": 0.7,
			  "queries": [
				{
				  "multi_match": {
					"type": "cross_fields",
					"query": %s,
					"analyzer": "whitespace",
					"auto_generate_synonyms_phrase_query": false,
					"operator": "and",
					"_name": %s,
					"fields": [
					  "package_attr_name^9",
					  "package_attr_name.*^5.3999999999999995",
					  "package_programs^9",
					  "package_programs.*^5.3999999999999995",
					  "package_pname^6",
					  "package_pname.*^3.5999999999999996",
					  "package_description^1.3",
					  "package_description.*^0.78",
					  "package_longDescription^1",
					  "package_longDescription.*^0.6",
					  "flake_name^0.5",
					  "flake_name.*^0.3"
					]
				  }
				},
				{
				  "wildcard": {
					"package_attr_name": {
					  "value": %s,
					  "case_insensitive": true
					}
				  }
				}
			  ]
			}
		  }
		]
	  }
	}
  }
`
)
