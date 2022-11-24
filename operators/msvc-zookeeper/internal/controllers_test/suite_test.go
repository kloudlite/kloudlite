package controllers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	zookeeperMsvcv1 "operators.kloudlite.io/apis/zookeeper.msvc/v1"
	testlib "operators.kloudlite.io/test-lib"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var schemes = testlib.AddToSchemes(zookeeperMsvcv1.AddToScheme)
var _ = testlib.PreSuite(schemes)

var _ = testlib.PostSuite()
