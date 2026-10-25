package main_test

import (
	"path/filepath"
	"sort"
	"testing"
)

// Preserve the complete byte snapshot, including Git and dependency state.
// Report names only: fixture contents do not belong in failure diagnostics.
func assertCandidateSnapshotUnchanged(t *testing.T, root string, before, after map[string]string) {
	t.Helper()
	var changed []string
	for path, original := range before {
		current, exists := after[path]
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			changed = append(changed, "removed "+relative)
		} else if current != original {
			changed = append(changed, "changed "+relative)
		}
	}
	for path := range after {
		if _, existed := before[path]; !existed {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			changed = append(changed, "added "+relative)
		}
	}
	if len(changed) != 0 {
		sort.Strings(changed)
		t.Fatalf("collector changed original Git state or sources: %q", changed)
	}
}
