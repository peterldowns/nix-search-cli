package nixsearch

import "fmt"

// Response is the format for an ElasticSearch API response.
// If the request was successful, only `Hits` will be populated.
// if the request failed, `Error` and `Status` will both be set, and `Hits` will be empty.
type Response struct {
	Error  *Error `json:"error"`
	Status *int   `json:"status"`
	Hits   struct {
		Hits []Hit `json:"hits"`
	} `json:"hits"`
}

type Error struct {
	Type         string `json:"type"`
	Reason       string `json:"reason"`
	ResourceType string `json:"resource.type"`
	ResourceID   string `json:"resource.id"`
}

func (e Error) Error() string {
	return fmt.Sprintf("API failure[%s](%s=%s): %s", e.Type, e.ResourceType, e.ResourceID, e.Reason)
}

type Hit struct {
	ID      string  `json:"_id"`
	Package Package `json:"_source"`
}

type License struct {
	FullName string `json:"fullName"`
	URL      string `json:"url"`
}

type FlakeResolved struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	URL   string `json:"url"`
}
type Package struct {
	Type             string        `json:"type"`
	Name             string        `json:"package_pname"`
	AttrName         string        `json:"package_attr_name"`
	AttrSet          string        `json:"package_attr_set"`
	Outputs          []string      `json:"package_outputs"`
	Description      string        `json:"package_description"`
	Programs         []string      `json:"package_programs"`
	Homepage         []string      `json:"package_homepage"`
	Version          string        `json:"package_pversion"`
	Platforms        []string      `json:"package_platforms"`
	Position         string        `json:"package_position"`
	Licenses         []License     `json:"package_license"`
	FlakeName        string        `json:"flake_name"`
	FlakeDescription string        `json:"flake_description"`
	FlakeResolved    FlakeResolved `json:"flake_resolved"`
}

func (p Package) IsFlake() bool {
	return p.FlakeResolved.Type != ""
}
