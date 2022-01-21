package pipelines_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKfp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kfp Suite")
}
