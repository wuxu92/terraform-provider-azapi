package storage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestStorageAcceptance is the Ginkgo entry point for the Microsoft.Storage native
// static-resource acceptance suite. Every resource in this package contributes
// Describe blocks (in its <resource>_test.go file) that this bootstrap runs. Ginkgo
// allows only one RunSpecs per test binary, so each generated service package owns
// exactly one of these. With TF_ACC unset every scenario skips, so `go test` stays
// fast and offline.
func TestStorageAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Microsoft.Storage native acceptance suite")
}
