// Package githubversions provides functions to interact with GitHub's API to retrieve version tags from a repository.
package githubversions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

// Tag represents a GitHub tag
type Tag struct {
	Name       string `json:"name"`
	Commit     Commit `json:"commit"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	NodeID     string `json:"node_id"`
}

// Commit represents a GitHub commit
type Commit struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

// GetSemverTags retrieves all version tags from a GitHub repository that follow semver convention
func GetSemverTags(ctx context.Context, owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid rate limiting issues
	req.Header.Set("User-Agent", "clyde-githubversions")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var semverTags []string
	for _, tag := range tags {
		// Add 'v' prefix if not present for semver validation
		version := tag.Name
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}

		// Check if the version follows semver convention
		if semver.IsValid(version) {
			semverTags = append(semverTags, tag.Name)
		}
	}

	// Sort tags by semver order (newest first)
	sort.Slice(semverTags, func(i, j int) bool {
		vi := semverTags[i]
		vj := semverTags[j]

		// Add 'v' prefix for comparison
		if !strings.HasPrefix(vi, "v") {
			vi = "v" + vi
		}
		if !strings.HasPrefix(vj, "v") {
			vj = "v" + vj
		}

		// Use semver.Compare if both are valid semver, otherwise use string comparison
		if semver.IsValid(vi) && semver.IsValid(vj) {
			return semver.Compare(vi, vj) > 0
		}

		// Fallback to string comparison for versions that don't pass strict semver validation
		return vi > vj
	})

	return semverTags, nil
}

// GetLatestSemverTag retrieves the latest version tag from a GitHub repository that follows semver convention
func GetLatestSemverTag(ctx context.Context, owner, repo string) (string, error) {
	tags, err := GetSemverTags(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no semver tags found for repository %s/%s", owner, repo)
	}

	return tags[0], nil
}

// GetLatestStableSemverTag retrieves the latest stable version tag from a GitHub repository that follows semver convention
func GetLatestStableSemverTag(ctx context.Context, owner, repo string) (string, error) {
	tags, err := GetSemverTags(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no semver tags found for repository %s/%s", owner, repo)
	}
	for _, tag := range tags {
		if semver.IsValid(tag) && semver.Prerelease(tag) == "" {
			return tag, nil
		}
	}

	return "", fmt.Errorf("no stable semver tags found for repository %s/%s", owner, repo)
}

// GetSemverTagsWithLimit retrieves version tags with a limit on the number of results
func GetSemverTagsWithLimit(ctx context.Context, owner, repo string, limit int) ([]string, error) {
	tags, err := GetSemverTags(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(tags) > limit {
		tags = tags[:limit]
	}

	return tags, nil
}
