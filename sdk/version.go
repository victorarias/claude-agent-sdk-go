package sdk

import (
	"fmt"
	"strconv"
	"strings"
)

// parseVersion parses a semantic version string.
func parseVersion(version string) ([3]int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid version format: %s", version)
	}

	var result [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return [3]int{}, fmt.Errorf("invalid version component: %s", part)
		}
		result[i] = n
	}

	return result, nil
}

// compareVersions compares two version strings.
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareVersions(a, b string) int {
	va, err := parseVersion(a)
	if err != nil {
		return -1
	}
	vb, err := parseVersion(b)
	if err != nil {
		return 1
	}

	for i := 0; i < 3; i++ {
		if va[i] < vb[i] {
			return -1
		}
		if va[i] > vb[i] {
			return 1
		}
	}
	return 0
}

// checkMinimumVersion validates that the CLI version meets minimum requirements.
func checkMinimumVersion(version string) error {
	if compareVersions(version, MinimumCLIVersion) < 0 {
		return &CLIVersionError{
			InstalledVersion: version,
			MinimumVersion:   MinimumCLIVersion,
		}
	}
	return nil
}
