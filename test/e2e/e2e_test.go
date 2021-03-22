package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	k8svizBinName = "k8sviz"
)

func getk8sviz() string {
	wd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	topDir := filepath.Dir(filepath.Dir(wd))
	return filepath.Join(topDir, "bin", k8svizBinName)
}

func logf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(GinkgoWriter, format+"\n", args...)
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("E2e", func() {
	It("Should create proper dot file with k8sviz for the created resources", func() {
		By("Creating resources")

		By("Creating dot file with k8sviz")
		k8sviz := getk8sviz()
		logf("%s", k8sviz)

		var stdout, stderr bytes.Buffer
		cmd := exec.Command("ls", "-lh", k8sviz)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		logf("Showing current directory: stdout: %s, stderr: %s",
			stdout.String(), stderr.String())

		By("Verifying created dot file")
	})
})
