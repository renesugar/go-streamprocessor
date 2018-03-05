package testutils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"
)

var validIntegrationTestName = regexp.MustCompile(`^(Test|Benchmark)(Integration)?.+$`)
var validRandomName = regexp.MustCompile(`[a-zA-Z0-9_]+`)

// Integration skips the test if the `-short` flag is specified. It also
// enforces the test name to starts with `Integration`, this allows to run
// _only_ integration tests using `-run '^TestIntegration'` or
// `-bench '^BenchmarkIntegration`.
func Integration(tb testing.TB) {
	tb.Helper()

	match := validIntegrationTestName.FindStringSubmatch(tb.Name())
	if len(match) < 3 || match[2] == "" {
		tb.Fatalf("integration test name %s does not start with %sIntegration", tb.Name(), match[1])
	}

	if testing.Short() {
		tb.Skip("integration test skipped due to -short")
	}
}

// Random returns a random string, to use during testing.
func Random(tb testing.TB) string {
	tb.Helper()

	rand.Seed(time.Now().Unix())
	name := strings.Replace(tb.Name(), "/", "_", -1)
	name = validRandomName.FindString(name)
	return fmt.Sprintf("%s-%d", name, rand.Intn(1000000))
}