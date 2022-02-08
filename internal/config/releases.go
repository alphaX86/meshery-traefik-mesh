package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"

	"github.com/layer5io/meshery-adapter-library/adapter"
)

// Release is used to save the release informations
type Release struct {
	ID      int             `json:"id,omitempty"`
	TagName string          `json:"tag_name,omitempty"`
	Name    adapter.Version `json:"name,omitempty"`
	Draft   bool            `json:"draft,omitempty"`
	Assets  []*Asset        `json:"assets,omitempty"`
}

// Asset describes the github release asset object
type Asset struct {
	Name        string `json:"name,omitempty"`
	State       string `json:"state,omitempty"`
	DownloadURL string `json:"browser_download_url,omitempty"`
}

// getLatestReleaseNames returns the names of the latest releases
// limited by the "limit" parameter. It filters out all the rc
// releases and sorts the result lexographically (descending)
func getLatestReleaseNames(limit int) ([]adapter.Version, error) {
	releases, err := GetLatestReleases()
	if err != nil {
		return []adapter.Version{}, ErrGetLatestReleaseNames(err)
	}

	// Filter out the rc releases
	result := make([]adapter.Version, limit)
	r, err := regexp.Compile(`\d+(\.\d+){2,}$`)
	if err != nil {
		return []adapter.Version{}, ErrGetLatestReleaseNames(err)
	}

	for _, release := range releases {
		versionStr := string(release.Name)
		if r.MatchString(versionStr) {
			result = append(result, adapter.Version(versionStr))
		}
	}

	// Sort the result
	sort.Slice(result, func(i, j int) bool {
		return result[i] > result[j]
	})

	if limit > len(result) {
		limit = len(result)
	}

	return result[:limit], nil
}

// GetLatestReleases fetches the latest releases from the traefik mesh repository
func GetLatestReleases() ([]*Release, error) {
	// Making the results to 10 to avoid fetching lot of releases (to avoid CWE-88)
	const releaseAPIURL = "https://api.github.com/repos/traefik/mesh/releases?per_page=10"
	// We need a variable url here hence using nosec
	// #nosec
	resp, err := http.Get(releaseAPIURL)
	if err != nil {
		return []*Release{}, ErrGetLatestReleases(err)
	}

	if resp.StatusCode != http.StatusOK {
		return []*Release{}, ErrGetLatestReleases(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []*Release{}, ErrGetLatestReleases(err)
	}

	var releaseList []*Release

	if err = json.Unmarshal(body, &releaseList); err != nil {
		return []*Release{}, ErrGetLatestReleases(err)
	}

	if err = resp.Body.Close(); err != nil {
		return []*Release{}, ErrGetLatestReleases(err)
	}

	return releaseList, nil
}
