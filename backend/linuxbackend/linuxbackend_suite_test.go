package linuxbackend_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLinuxbackend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linuxbackend Suite")
}
