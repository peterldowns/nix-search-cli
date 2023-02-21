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

type Input struct {
	Channel  string
	Default  string
	Program  string
	Advanced string
	Name     string
}

func (c Client) Search(ctx context.Context, input Input) ([]Package, error) {
	req, err := buildRequest(ctx, input)
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

func buildRequest(ctx context.Context, input Input) (*http.Request, error) {
	url := formatURL(input.Channel)
	payload, err := formatQuery(input)
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

func formatQuery(input Input) (string, error) {
	var queries []string
	if input.Default != "" {
		q, err := DefaultQuery(input.Default)
		if err != nil {
			return "", err
		}
		queries = append(queries, q)
	}
	if input.Name != "" {
		q, err := AttrQuery(input.Name)
		if err != nil {
			return "", err
		}
		queries = append(queries, q)
	}
	if input.Program != "" {
		q, err := ProgramQuery(input.Program)
		if err != nil {
			return "", err
		}
		queries = append(queries, q)
	}
	if input.Advanced != "" {
		q, err := AdvancedQuery(input.Advanced)
		if err != nil {
			return "", err
		}
		queries = append(queries, q)
	}
	query := strings.Join(queries, ", ")
	return fmt.Sprintf(payloadTemplate, query), nil
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
	"query": {
		"bool": {
			"must": [
				%s
			]
		}
	}
}
`
)

/*
nix-search this is the default query works just like the website
nix-search -q this is the default query works just like the website
nix-search --program gcloud # substring match on program entry
nix-search --attr  name # substring match on attribute name
*/
func ProgramQuery(query string) (string, error) {
	encQuery, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
{
	"wildcard": {
		"package_programs": {
			"case_insensitive": true,
			"value": %s
		}
	}
}
	`, encQuery), nil
}

func AdvancedQuery(query string) (string, error) {
	encQuery, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
{
	"query_string": {
		"query": %s
	}
}
	`, encQuery), nil
}

func AttrQuery(query string) (string, error) {
	encQuery, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
{
	"wildcard": {
		"package_attr_name": {
			"case_insensitive": true,
			"value": %s
		}
	}
}
	`, encQuery), nil
}

func DefaultQuery(query string) (string, error) {
	matchName := "multi_match_" + strings.ReplaceAll(query, " ", "_")
	value := fmt.Sprintf("*%s*", query)

	encQuery, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	encMatchName, err := json.Marshal(matchName)
	if err != nil {
		return "", err
	}
	encValue, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	tpl := `
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
						"package_longDescription.*^0.6"
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
`
	return fmt.Sprintf(tpl, encQuery, encMatchName, encValue), nil
}
