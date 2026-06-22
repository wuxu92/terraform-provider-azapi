package nativeacc

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestNativeAcceptance is the single Ginkgo entry point for the native static
// resource acceptance suite. Ginkgo allows only one RunSpecs per test binary, so
// every resource contributes Describe blocks (in its own <resource>_test.go file)
// that this bootstrap runs. With TF_ACC unset, every scenario skips, so `go test`
// stays fast and offline.
func TestNativeAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "native static resources Acceptance Suite")
}
