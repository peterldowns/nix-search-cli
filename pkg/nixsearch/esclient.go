package nixsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
)

// All constants are taken from the upstream repository
// https://github.com/NixOS/nixos-search/blob/main/frontend/src/index.js
const (
	ElasticSearchUsername    = "aWVSALXpZv"
	ElasticSearchPassword    = "X8gPHnzL52wFEekuxsfQ9cSh"
	ElasticSearchIndexPrefix = "latest-37-"
	ElasticSearchURLTemplate = `https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net:443/%s/_search`
)

type ElasticSearchClient struct {
	HTTPClient *http.Client
}

func NewElasticSearchClient() (*ElasticSearchClient, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = nil

	return &ElasticSearchClient{
		HTTPClient: retryClient.StandardClient(),
	}, nil
}

func (c ElasticSearchClient) Search(ctx context.Context, input Query) ([]Package, error) {
	req, err := newRequest(ctx, input)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	packages, err := readResponse(resp)
	if err != nil {
		return nil, err
	}

	var out []Package
	for _, p := range packages {
		if p.Type != "package" {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func newRequest(ctx context.Context, input Query) (*http.Request, error) {
	index := ""
	if input.Flakes {
		index = ElasticSearchIndexPrefix + "group-manual"
	} else {
		index = ElasticSearchIndexPrefix + url.QueryEscape("nixos-"+input.Channel)
	}
	url := fmt.Sprintf(ElasticSearchURLTemplate, index)
	payload, err := input.Payload()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(ElasticSearchUsername, ElasticSearchPassword)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func readResponse(resp *http.Response) ([]Package, error) {
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
