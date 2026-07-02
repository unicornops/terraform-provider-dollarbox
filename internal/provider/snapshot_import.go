package provider

import (
	"fmt"
	"strconv"
	"strings"
)

func parseSnapshotImportID(importID string, expectedParts int) ([]string, error) {
	parts := strings.Split(strings.TrimSpace(importID), "/")
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("expected %d slash-separated parts, got %q", expectedParts, importID)
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return nil, fmt.Errorf("import ID parts must not be empty: %q", importID)
		}
	}
	if _, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
		return nil, fmt.Errorf("namespace ID must be an integer in import ID %q: %w", importID, err)
	}
	return parts, nil
}
