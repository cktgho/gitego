// utils/main_test.go
package utils_test

import (
	"os"
	"testing"
)

// TestMain is the entry point for tests in this package. It allows us to
// check for the helper process environment variable.
func TestMain(m *testing.M) {
	// The TestHelperProcess function will handle the mock execution.
	// This setup allows the other tests to run normally.
	os.Exit(m.Run())
}
