package web_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestWebAcceptance is the Ginkgo entry point for the Microsoft.Web native
// static-resource acceptance suite.
func TestWebAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Microsoft.Web native acceptance suite")
}
