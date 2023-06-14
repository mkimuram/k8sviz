// SPDX-FileCopyrightText: 2019 - 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkimuram/k8sviz/pkg/graph"
	"github.com/mkimuram/k8sviz/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	defaultNamespace   		= "default"
	defaultOutFile     		= "k8sviz.out"
	defaultOutType     		= "dot"
	defaultLabelSelector	= ""
	defaultFieldSelector	= ""
	descNamespaceOpt   		= "namespace to visualize"
	descOutFileOpt     		= "output filename"
	descOutTypeOpt     		= "type of output"
	descShortOptSuffix 		= " (shorthand)"
	descLabelSelector		= "label selector"
	descFieldSelector		= "field selector"
)

var (
	clientset *kubernetes.Clientset
	dir       		string
	// Flags
	namespace 		string
	outFile   		string
	outType   		string
	labelSelector	string
	fieldSelector	string
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
	flag.StringVar(&labelSelector, "l", defaultLabelSelector, descLabelSelector)
	flag.StringVar(&labelSelector, "selector", defaultLabelSelector, descLabelSelector)
	flag.StringVar(&fieldSelector, "field-selector", defaultFieldSelector, descFieldSelector)
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
	_, err = clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
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
	res, err := resources.NewResources(clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get k8s resources: %v\n", err)
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			fmt.Fprintf(os.Stderr, "k8sviz 0.3.3 or later only support k8s 1.21 or later.\n")
			fmt.Fprintf(os.Stderr, "If you are using older k8s cluster, try k8sviz 0.3.2 or earlier.\n")
		}
		os.Exit(1)
	}

	g := graph.NewGraph(res, dir)

	if outType == "dot" {
		if err := g.WriteDotFile(outFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output %q file with format %q for namespace %q: %v\n", outFile, outType, namespace, err)
			os.Exit(1)
		}
	} else {
		if err := g.PlotDotFile(outFile, outType); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output %q file with format %q for namespace %q: %v\n", outFile, outType, namespace, err)
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
