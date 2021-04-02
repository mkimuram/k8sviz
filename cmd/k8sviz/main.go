// SPDX-FileCopyrightText: 2019 - 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"github.com/mkimuram/k8sviz/pkg/graph"
	"github.com/mkimuram/k8sviz/pkg/resources"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	defaultNamespace   = "namespace"
	defaultOutDir      = "out"
	defaultOutType     = "dot"
	descNamespaceOpt   = "namespace to visualize"
	descOutDirOpt      = "output dir"
	descOutTypeOpt     = "type of output"
	descShortOptSuffix = " (shorthand)"
)

var (
	clientset *kubernetes.Clientset
	dir       string
	// Flags
	namespace string
	outDir    string
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
	flag.StringVar(&outDir, "outDir", defaultOutDir, descOutDirOpt)
	flag.StringVar(&outDir, "o", defaultOutDir, descOutDirOpt+descShortOptSuffix)
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
	//if !strings.Contains(namespace,",") {
	//	_, err = clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	//	if err != nil {
	//		fmt.Fprintf(os.Stderr, "Failed to get namespace %q: %v\n", namespace, err)
	//		os.Exit(1)
	//	}
	//}

	dir, err = getBinDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find the directory of this command: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// Get all resources in the namespace
	words := strings.Split(namespace, ",")
	for _, ns := range words {
		res, err := resources.NewResources(clientset, ns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get k8s resources: %v\n", err)
			os.Exit(1)
		}

		g := graph.NewGraph(res, dir)
		if outType == "dot" {
			if err := g.WriteDotFile(outDir + "/" + ns + "." + outType); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to output %q file with format %q for namespace %q: %v\n", outDir, outType, namespace, err)
				os.Exit(1)
			}
		} else {
			if err := g.PlotDotFile(outDir+"/"+ns+"."+outType, outType); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to output %q file with format %q for namespace %q: %v\n", outDir, outType, namespace, err)
				os.Exit(1)
			}
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
