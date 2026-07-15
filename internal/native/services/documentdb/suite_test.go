package documentdb_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestDocumentDBAcceptance is the Ginkgo entry point for the Microsoft.DocumentDB
// native static-resource acceptance suite. Every resource in this package contributes
// Describe blocks (in its <resource>_test.go file) that this bootstrap runs. Ginkgo
// allows only one RunSpecs per test binary, so each generated service package owns
// exactly one of these. With TF_ACC unset every scenario skips, so `go test` stays
// fast and offline.
func TestDocumentDBAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Microsoft.DocumentDB native acceptance suite")
}
