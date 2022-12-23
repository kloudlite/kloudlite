package controllers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	testlib "operators.kloudlite.io/testing"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var schemes = testlib.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
var _ = testlib.PreSuite(schemes)

var _ = testlib.PostSuite()
