package githubversions

import (
	"context"
	"testing"
)

func TestGetSemverTags(t *testing.T) {
	ctx := context.Background()

	// Test with a well-known repository that has semver tags
	owner := "projectcalico"
	repo := "calico"

	tags, err := GetSemverTags(ctx, owner, repo)
	if err != nil {
		t.Fatalf("Failed to get semver tags: %v", err)
	}

	if len(tags) == 0 {
		t.Log("No semver tags found (this might be expected for some repositories)")
		return
	}

	t.Logf("Found %d semver tags", len(tags))
	for i, tag := range tags {
		if i >= 5 { // Only show first 5 tags
			t.Logf("... and %d more tags", len(tags)-5)
			break
		}
		t.Logf("Tag %d: %s", i+1, tag)
	}
}

func TestGetLatestSemverTag(t *testing.T) {
	ctx := context.Background()

	// Test with a well-known repository that has semver tags
	owner := "projectcalico"
	repo := "calico"

	latest, err := GetLatestSemverTag(ctx, owner, repo)
	if err != nil {
		t.Fatalf("Failed to get latest semver tag: %v", err)
	}

	t.Logf("Latest Calico semver tag: %s", latest)
}

func TestGetLatestStableSemverTag(t *testing.T) {
	ctx := context.Background()

	// Test with a well-known repository that has semver tags
	owner := "projectcalico"
	repo := "calico"

	latest, err := GetLatestStableSemverTag(ctx, owner, repo)
	if err != nil {
		t.Fatalf("Failed to get latest stable semver tag: %v", err)
	}

	t.Logf("Latest Stable Calico semver tag: %s", latest)
}

func TestGetSemverTagsWithLimit(t *testing.T) {
	ctx := context.Background()

	// Test with a well-known repository that has semver tags
	owner := "projectcalico"
	repo := "calico"

	tags, err := GetSemverTagsWithLimit(ctx, owner, repo, 3)
	if err != nil {
		t.Fatalf("Failed to get semver tags with limit: %v", err)
	}

	if len(tags) > 3 {
		t.Errorf("Expected at most 3 tags, got %d", len(tags))
	}

	t.Logf("Found %d tags (limited to 3)", len(tags))
	for i, tag := range tags {
		t.Logf("Tag %d: %s", i+1, tag)
	}
}
