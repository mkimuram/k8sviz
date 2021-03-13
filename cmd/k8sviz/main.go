// SPDX-FileCopyrightText: 2019 - 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mkimuram/k8sviz/pkg/graph"
	"github.com/mkimuram/k8sviz/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	defaultNamespace   = "namespace"
	defaultOutFile     = "k8sviz.out"
	defaultOutType     = "dot"
	descNamespaceOpt   = "namespace to visualize"
	descOutFileOpt     = "output filename"
	descOutTypeOpt     = "type of output"
	descShortOptSuffix = " (shorthand)"
)

var (
	clientset *kubernetes.Clientset
	dir       string
	// Flags
	namespace string
	outFile   string
	outType   string
)

func init() {
	var (
		err        error
		kubeconfig string
	)
	if home := os.Getenv("HOME"); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&namespace, "namespace", defaultNamespace, descNamespaceOpt)
	flag.StringVar(&namespace, "n", defaultNamespace, descNamespaceOpt+descShortOptSuffix)
	flag.StringVar(&outFile, "outfile", defaultOutFile, descOutFileOpt)
	flag.StringVar(&outFile, "o", defaultOutFile, descOutFileOpt+descShortOptSuffix)
	flag.StringVar(&outType, "type", defaultOutType, descOutTypeOpt)
	flag.StringVar(&outType, "t", defaultOutType, descOutTypeOpt+descShortOptSuffix)
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build config from %q: %v\n", kubeconfig, err)
		os.Exit(1)
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client from %q: %v\n", kubeconfig, err)
		os.Exit(1)
	}

	// test connectivity for k8s cluster and the namespace
	_, err = clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get namespace %q: %v\n", namespace, err)
		os.Exit(1)
	}

	dir, err = getBinDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find the directory of this command: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// Get all resources in the namespace
	res := resources.NewResources(clientset, namespace)

	if outType == "dot" {
		if err := graph.WriteDotFile(res, dir, namespace, outFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output dot file for namespace %q: %v\n", namespace, err)
			os.Exit(1)
		}
	} else {
		if err := graph.PlotDotFile(res, dir, namespace, outFile, outType); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output %s file for namespace %q: %v\n", outType, namespace, err)
			os.Exit(1)
		}
	}
}

func getBinDir() (string, error) {
	s, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(s), nil
}
