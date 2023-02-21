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
	Channel    string
	Default    string
	Program    string
	Advanced   string
	Name       string
	MaxResults int
	Version    string
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
	// fmt.Println(payload)
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
	if input.Version != "" {
		q, err := VersionQuery(input.Version)
		if err != nil {
			return "", err
		}
		queries = append(queries, q)
	}
	query := strings.Join(queries, ", ")
	return fmt.Sprintf(payloadTemplate, input.MaxResults, query), nil
}

const (
	// https://github.com/NixOS/nixos-search/blob/main/frontend/src/index.js
	defaultUsername = "aWVSALXpZv"
	defaultPassword = "X8gPHnzL52wFEekuxsfQ9cSh"
	urlTemplate     = `https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net:443/latest-37-nixos-%s/_search`
	payloadTemplate = `
{
	"from": 0,
	"size": %d,
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

func VersionQuery(version string) (string, error) {
	wildcard, _ := json.Marshal(version + "*")
	encVersion, _ := json.Marshal(version)
	x := fmt.Sprintf(`
{
	"dis_max": {
		"tie_breaker": 0.7,
		"queries": [
			{
				"wildcard": {
					"package_pversion": {
						"value": %s
					}
				}
			},
			{
				"match": {
					"package_pversion": %s
				}
			}
		]
	}
}
	`, wildcard, encVersion)
	return x, nil
}

func ProgramQuery(query string) (string, error) {
	wildcard, _ := json.Marshal(query + "*")
	encQuery, _ := json.Marshal(query)
	x := fmt.Sprintf(`
{
	"dis_max": {
		"tie_breaker": 0.7,
		"queries": [
			{
				"wildcard": {
					"package_programs": {
						"value": %s
					}
				}
			},
			{
				"match": {
					"package_programs": %s
				}
			}
		]
	}
}
	`, wildcard, encQuery)
	return x, nil
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
	wildcard, _ := json.Marshal(query + "*")
	encQuery, _ := json.Marshal(query)
	x := fmt.Sprintf(`
{
	"dis_max": {
		"tie_breaker": 0.7,
		"queries": [
			{
				"wildcard": {
					"package_attr_name": {
						"value": %s
					}
				}
			},
			{
				"match": {
					"package_attr_name": %s
				}
			}
		]
	}
}
	`, wildcard, encQuery)
	return x, nil
}

func DefaultQuery(query string) (string, error) {
	matchName := "multi_match_" + strings.ReplaceAll(query, " ", "_")
	encQuery, _ := json.Marshal(query)
	encMatchName, _ := json.Marshal(matchName)

	var queries []string
	multiMatch := fmt.Sprintf(`
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
						"package_pversion^1.3",
						"package_pversion.*^0.78",
						"package_longDescription^1",
						"package_longDescription.*^0.6"
					]
				}
			}
			`, encQuery, encMatchName)
	queries = append(queries, multiMatch)
	for _, term := range strings.Split(query, " ") {
		enc, _ := json.Marshal("*" + term + "*")
		wildcard := fmt.Sprintf(`
			{
				"wildcard": {
					"package_attr_name": {
						"value": %s,
						"case_insensitive": true
					}
				}
			}
		`, enc)
		queries = append(queries, wildcard)
	}
	tpl := `
{
	"dis_max": {
		"tie_breaker": 0.7,
		"queries": [%s]
	}
}
`
	return fmt.Sprintf(tpl, strings.Join(queries, ", ")), nil
}
