package integration_test

import "testing"

// skipIfRaceDetector skips the test when the race detector is enabled.
// This is used to skip tests that trigger a known data race in hashicorp/go-plugin v1.7.0.
// The race is internal to Client.Start() and cannot be fixed without upstream changes.
func skipIfRaceDetector(t *testing.T) {
	t.Helper()
	if raceEnabled {
		t.Skip("Skipping due to known data race in hashicorp/go-plugin v1.7.0")
	}
}
