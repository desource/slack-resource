package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSlackResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "slack-resource test suite")
}

var _ = Describe("check", func() {
	It("should should ignore check", func() {

		Expect(false).To(Equal(true))

	})
})
