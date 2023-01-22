package controllers_test

import (
	"testing"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	testlib "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var schemes = testlib.AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
var _ = testlib.PreSuite(schemes)

var _ = testlib.PostSuite()
