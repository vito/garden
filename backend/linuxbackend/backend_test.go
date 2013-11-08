package linuxbackend_test

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/garden/backend"
	"github.com/vito/garden/backend/linuxbackend"
)

var _ = Describe("Creating containers", func() {
	var linuxBackend *linuxbackend.LinuxBackend
	var depotPath string

	BeforeEach(func() {
		_, currentFile, _, _ := runtime.Caller(0)
		currentDirectory := path.Dir(currentFile)

		rootDirectory := path.Join(
			path.Dir(path.Dir(currentDirectory)),
			"warden",
			"warden",
			"root",
			"linux",
		)

		tmpdir, err := ioutil.TempDir(os.TempDir(), "depot")
		Expect(err).ToNot(HaveOccured())

		depotPath = tmpdir

		linuxBackend = linuxbackend.New(rootDirectory, depotPath)
	})

	It("allocates a user and network for the container", func() {

	})

	It("bind-mounts the given bind-mounts", func() {

	})

	It("starts wshd", func() {

	})
})
