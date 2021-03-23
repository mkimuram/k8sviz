package e2e_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	k8svizBinName = "k8sviz"
)

func getTopDir() string {
	// current working directory should be {top_dir}/test/e2e
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current working directory")
		os.Exit(1)
	}
	// {top_dir}/test/e2e/../.. will be {top_dir}
	return filepath.Dir(filepath.Dir(wd))
}

func getDefaultBinPath() string {
	// returns {top_dir}/bin/k8sviz
	return filepath.Join(getTopDir(), "bin", k8svizBinName)
}

func getDefaultTestDataPath() string {
	// returns {top_dir}/test/data
	return filepath.Join(getTopDir(), "test", "data")
}

func logf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(GinkgoWriter, format+"\n", args...)
	Expect(err).NotTo(HaveOccurred())
}

var (
	kubeconfig  string
	testDir     string
	testBin     string
	testDataDir string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Path to kubeconfig")
	flag.StringVar(&testDir, "testdir", "/tmp", "Path to testdir")
	flag.StringVar(&testBin, "testbin", getDefaultBinPath(), "Path to the binary to be tested")
	flag.StringVar(&testDataDir, "testdata", getDefaultTestDataPath(), "Path that has test data")
}

func runK8sviz(bin, config, namespace, outType, outFile string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, "-kubeconfig", config, "-namespace", namespace, "-type", outType, "-outfile", outFile)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func verifyFileType(name, fileType string) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("file", name)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred(), "running file command failed: stdout: %s, stderr: %s, err: %v", stdout.String(), stderr.String(), err)

	logf("Result of file command for %q: %s", name, stdout.String())

	switch fileType {
	case "dot":
		Expect(stdout.String()).Should(ContainSubstring("ASCII text"))
	case "png":
		Expect(stdout.String()).Should(ContainSubstring("PNG image data"))
	}
}

func getClientSet(config string) *kubernetes.Clientset {
	conf, err := clientcmd.BuildConfigFromFlags("", config)
	Expect(err).NotTo(HaveOccurred(), "Failed to build config from kubeconfig %q, err: %v", config, err)

	cs, err := kubernetes.NewForConfig(conf)
	Expect(err).NotTo(HaveOccurred(), "Failed to create clientset from config, err: %v", err)

	return cs
}

func getObjFromFile(path string) runtime.Object {
	yaml, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred(), "Failed to read from file %q, err: %v", path, err)

	sch := runtime.NewScheme()
	_ = appsv1.AddToScheme(sch)
	_ = batchv1.AddToScheme(sch)
	_ = corev1.AddToScheme(sch)
	_ = v1beta1.AddToScheme(sch)

	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	obj, _, err := decode(yaml, nil, nil)
	Expect(err).NotTo(HaveOccurred(), "Failed to decode %q: %v", path, err)
	return obj
}

func createFromFile(cs *kubernetes.Clientset, path, namespace string) {
	obj := getObjFromFile(path)

	switch o := obj.(type) {
	case *corev1.Pod:
		_, err := cs.CoreV1().Pods(namespace).Create(o)
		Expect(err).NotTo(HaveOccurred(), "Failed to create pod %v: %v", o, err)
	}
}

func deleteFromFile(cs *kubernetes.Clientset, path, namespace string) {
	obj := getObjFromFile(path)

	switch o := obj.(type) {
	case *corev1.Pod:
		err := cs.CoreV1().Pods(namespace).Delete(o.GetName(), &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred(), "Failed to create pod %v: %v", o, err)
	}
}

var _ = Describe("E2e", func() {
	It("Should create proper dot file with k8sviz for the created resources", func() {
		logf("Running test with args: kubeconfig: %s, testDir: %s, testBin: %s, testDataDir: %s", kubeconfig, testDir, testBin, testDataDir)
		By("Creating resources")
		cs := getClientSet(kubeconfig)
		yaml := filepath.Join(testDataDir, "pod.yaml")
		createFromFile(cs, yaml, "default")

		By("Creating dot file with k8sviz")
		dotFile := filepath.Join(testDir, "foo.dot")

		stdout, stderr, err := runK8sviz(testBin, kubeconfig, "kube-system", "dot", dotFile)
		Expect(err).NotTo(HaveOccurred(), "running k8sviz command failed: stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		By("Verifying created dot file")
		verifyFileType(dotFile, "dot")
		dot, err := ioutil.ReadFile(dotFile)
		Expect(err).NotTo(HaveOccurred())
		logf("dot:\n%s", string(dot))

		// TODO: More checks like comparing with golden files

		By("Creating png file with k8sviz")
		pngFile := filepath.Join(testDir, "foo.png")

		stdout, stderr, err = runK8sviz(testBin, kubeconfig, "kube-system", "png", pngFile)
		Expect(err).NotTo(HaveOccurred(), "running k8sviz command failed: stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		By("Verifying created png file")
		verifyFileType(pngFile, "png")

		By("Deleting resources")
		deleteFromFile(cs, yaml, "default")
	})
})
