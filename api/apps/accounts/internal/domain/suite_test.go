package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestDomain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accounts Domain Suite")
}
