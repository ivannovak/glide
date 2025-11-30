package version

import (
	"testing"

	"github.com/ivannovak/glide/v2/tests/testutil"
)

// TestGetVersionString_TableDriven demonstrates the table test framework
// for testing version string formatting
func TestGetVersionString_TableDriven(t *testing.T) {
	// Store original value and restore after test
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	cases := []testutil.TestCase[string]{
		{
			Name:     "development version",
			Input:    "dev",
			Expected: "glide version dev (development build)",
		},
		{
			Name:     "release version v1.0.0",
			Input:    "v1.0.0",
			Expected: "glide version v1.0.0",
		},
		{
			Name:     "release version v2.3.1",
			Input:    "v2.3.1",
			Expected: "glide version v2.3.1",
		},
		{
			Name:     "pre-release version",
			Input:    "v1.0.0-beta.1",
			Expected: "glide version v1.0.0-beta.1",
		},
	}

	testutil.RunSimpleTableTests(t, cases, nil, func(_ *testing.T, versionInput string) string {
		// Set the Version for this test case
		Version = versionInput
		return GetVersionString()
	})
}

// TestGet_TableDriven demonstrates simple table testing
func TestGet_TableDriven(t *testing.T) {
	// Store original value and restore after test
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	cases := []testutil.TestCase[string]{
		{
			Name:     "version 1.0.0",
			Input:    "1.0.0",
			Expected: "1.0.0",
		},
		{
			Name:     "version dev",
			Input:    "dev",
			Expected: "dev",
		},
		{
			Name:     "version with v prefix",
			Input:    "v2.0.0",
			Expected: "v2.0.0",
		},
		{
			Name:     "empty version",
			Input:    "",
			Expected: "",
		},
	}

	testutil.RunSimpleTableTests(t, cases, nil, func(_ *testing.T, versionInput string) string {
		// Set the Version for this test case
		Version = versionInput
		return Get()
	})
}
