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

const (
	// Taken from the upstream repository
	// https://github.com/NixOS/nixos-search/blob/main/frontend/src/index.js
	ElasticSearchUsername    = "aWVSALXpZv"
	ElasticSearchPassword    = "X8gPHnzL52wFEekuxsfQ9cSh"
	ElasticSearchURLTemplate = `https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net:443/%s/_search`
	// See the list of available indexes at
	// https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net:443/_aliases
	// They're in the format "latest-<VERSION>-identifier", e.g.
	//   - latest-40-group-manual
	//   - latest-40-nixos-unstable
	//   - latest-40-nixos-22.11
	// As the indices are updated, the version number changes over time, and old
	// version numbers stop working because the related indices are deleted. The
	// upstream project does not create a version-number-less alias (although
	// they could easily) so we do the next best thing and use a wildcard prefix
	// for the version number. Experimentally, results are the same as before,
	// it doesn't matter that we're querying over multiple indices.
	ElasticSearchIndexPrefix = "latest-*-"
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

func (c ElasticSearchClient) Search(ctx context.Context, query Query) ([]Package, error) {
	req, err := newRequest(ctx, query)
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

func newRequest(ctx context.Context, query Query) (*http.Request, error) {
	index := ""
	if query.Flakes {
		index = ElasticSearchIndexPrefix + "group-manual"
	} else {
		index = ElasticSearchIndexPrefix + url.QueryEscape("nixos-"+query.Channel)
	}
	eurl := fmt.Sprintf(ElasticSearchURLTemplate, index)
	payload, err := query.Payload()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, eurl, bytes.NewReader(payload))
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
