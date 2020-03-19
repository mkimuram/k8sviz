package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/awalterschulze/gographviz"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	clusterPrefix      = "cluster_"
	rankPrefix         = "rank_"
	imageSuffix        = "-128.png"
	defaultNamespace   = "namespace"
	defaultOutFile     = "k8sviz.out"
	defaultOutType     = "dot"
	descNamespaceOpt   = "namespace to visualize"
	descOutFileOpt     = "output filename"
	descOutTypeOpt     = "type of output"
	descShortOptSuffix = " (shorthand)"
)

var (
	clientset       *kubernetes.Clientset
	dir             string
	resourceTypes   = []string{"deploy job", "sts ds rs", "pod", "pvc", "svc"}
	normalizedNames = map[string]string{
		"ns":     "namespace",
		"svc":    "service",
		"pvc":    "persistentvolumeclaim",
		"pod":    "po",
		"sts":    "statefulset",
		"ds":     "daemonset",
		"rs":     "replicaset",
		"deploy": "deployment",
		"job":    "job",
	}

	// Flags
	namespace string
	outFile   string
	outType   string
)

type resources struct {
	clientset *kubernetes.Clientset
	svcs      *corev1.ServiceList
	pvcs      *corev1.PersistentVolumeClaimList
	pods      *corev1.PodList
	stss      *appsv1.StatefulSetList
	dss       *appsv1.DaemonSetList
	rss       *appsv1.ReplicaSetList
	deploys   *appsv1.DeploymentList
	jobs      *batchv1.JobList
}

func newResources(clientset *kubernetes.Clientset, namespace string) *resources {
	var err error
	res := &resources{clientset: clientset}

	// service
	res.svcs, err = clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get services in namespace %q: %v\n", namespace, err)
	}

	// persistentvolumeclaim
	res.pvcs, err = clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get persistentVolumeClaims in namespace %q: %v\n", namespace, err)
	}

	// pod
	res.pods, err = clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get pods in namespace %q: %v\n", namespace, err)
	}

	// statefulset
	res.stss, err = clientset.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get statefulsets in namespace %q: %v\n", namespace, err)
	}

	// daemonset
	res.dss, err = clientset.AppsV1().DaemonSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get daemonsets in namespace %q: %v\n", namespace, err)
	}

	// replicaset
	res.rss, err = clientset.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get replicasets in namespace %q: %v\n", namespace, err)
	}

	// deployment
	res.deploys, err = clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get deployments in namespace %q: %v\n", namespace, err)
	}

	// job
	res.jobs, err = clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get jobs in namespace %q: %v\n", namespace, err)
	}

	return res
}

func (r *resources) getResourceNames(kind string) []string {
	names := []string{}

	switch kind {
	case "svc":
		for _, n := range r.svcs.Items {
			names = append(names, n.Name)
		}
	case "pvc":
		for _, n := range r.pvcs.Items {
			names = append(names, n.Name)
		}
	case "pod":
		for _, n := range r.pods.Items {
			names = append(names, n.Name)
		}
	case "sts":
		for _, n := range r.stss.Items {
			names = append(names, n.Name)
		}
	case "ds":
		for _, n := range r.dss.Items {
			names = append(names, n.Name)
		}
	case "rs":
		for _, n := range r.rss.Items {
			names = append(names, n.Name)
		}
	case "deploy":
		for _, n := range r.deploys.Items {
			names = append(names, n.Name)
		}
	case "job":
		for _, n := range r.jobs.Items {
			names = append(names, n.Name)
		}
	}

	return names
}

func (r *resources) hasResource(kind, name string) bool {
	for _, resName := range r.getResourceNames(kind) {
		if resName == name {
			return true
		}
	}
	return false
}

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
	if outType == "dot" {
		if err := writeDotFile(namespace, outFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output dot file for namespace %q: %v\n", namespace, err)
			os.Exit(1)
		}
	} else {
		if err := plotDotFile(namespace, outFile, outType); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output %s file for namespace %q: %v\n", outType, namespace, err)
			os.Exit(1)
		}
	}
}

func writeDotFile(namespace, outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(toDot(namespace)); err != nil {
		return err
	}

	return nil
}

func plotDotFile(namespace, outFile, outType string) error {
	cmd := exec.Command("dot", "-T"+outType, "-o", outFile)
	cmd.Stdin = strings.NewReader(toDot(namespace))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func toDot(namespace string) string {
	graph := gographviz.NewGraph()
	graph.SetDir(true)
	graph.SetName("G")
	graph.AddAttr("G", "rankdir", "TD")
	sname := toClusterName(namespace)
	graph.AddSubGraph("G", sname,
		map[string]string{"label": toClusterLabel(namespace), "labeljust": "l", "style": "dotted"})

	// Create subgraphs for resources to group by rank
	for r := 0; r < len(resourceTypes); r++ {
		graph.AddSubGraph(sname, toRankName(r),
			map[string]string{"rank": "same", "style": "invis"})
		// Put dummy invisible node to order ranks
		graph.AddNode(toRankName(r), toRankDummyNodeName(r),
			map[string]string{"style": "invis", "height": "0", "width": "0", "margin": "0"})
	}

	// Order ranks
	for r := 0; r < len(resourceTypes)-1; r++ {
		// Connect rth node and r+1th dummy node with invisible edge
		graph.AddEdge(toRankDummyNodeName(r), toRankDummyNodeName(r+1), true,
			map[string]string{"style": "invis"})
	}

	// Get all resources in the namespace
	res := newResources(clientset, namespace)

	// Put resources as Nodes in each rank of subgraph
	for r, rankRes := range resourceTypes {
		for _, resType := range strings.Fields(rankRes) {
			for _, name := range res.getResourceNames(resType) {
				graph.AddNode(toRankName(r), toResourceName(resType, name),
					map[string]string{"label": toResourceLabel(resType, name), "penwidth": "0"})
			}
		}
	}

	// Connect resources
	// Owner reference for pod
	for _, pod := range res.pods.Items {
		for _, ref := range pod.GetOwnerReferences() {
			ownerKind, err := normalizeResource(ref.Kind)
			if err != nil {
				// Skip resource that isn't available for this tool, like CRD
				continue
			}
			if !res.hasResource(ownerKind, ref.Name) {
				fmt.Fprintf(os.Stderr, "%s %s not found as a owner refernce for po %s\n", ownerKind, ref.Name, pod.Name)
				continue
			}
			graph.AddEdge(toResourceName(ownerKind, ref.Name), toResourceName("pod", pod.Name), true,
				map[string]string{"style": "dashed"})
		}
	}
	// Owner reference for rs
	for _, rs := range res.rss.Items {
		for _, ref := range rs.GetOwnerReferences() {
			ownerKind, err := normalizeResource(ref.Kind)
			if err != nil {
				// Skip resource that isn't available for this tool, like CRD
				continue
			}
			if !res.hasResource(ownerKind, ref.Name) {
				fmt.Fprintf(os.Stderr, "%s %s not found as a owner refernce for rs %s\n", ownerKind, ref.Name, rs.Name)
				continue
			}

			graph.AddEdge(toResourceName(ownerKind, ref.Name), toResourceName("rs", rs.Name), true,
				map[string]string{"style": "dashed"})
		}
	}

	// pvc and pod
	for _, pod := range res.pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.VolumeSource.PersistentVolumeClaim != nil {
				if !res.hasResource("pvc", vol.VolumeSource.PersistentVolumeClaim.ClaimName) {
					fmt.Fprintf(os.Stderr, "pvc %s not found as a volume for pod %s\n", vol.VolumeSource.PersistentVolumeClaim.ClaimName, pod.Name)
					continue
				}

				graph.AddEdge(toResourceName("pod", pod.Name), toResourceName("pvc", vol.VolumeSource.PersistentVolumeClaim.ClaimName), true,
					map[string]string{"dir": "none"})
			}
		}
	}

	// svc and pod
	for _, svc := range res.svcs.Items {
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		// Check if pod has all labels specified in svc.Spec.Selector
		for _, pod := range res.pods.Items {
			podLabel := pod.GetLabels()
			matched := true
			for selKey, selVal := range svc.Spec.Selector {
				val, ok := podLabel[selKey]
				if !ok || selVal != val {
					matched = false
					break
				}
			}

			if matched {
				graph.AddEdge(toResourceName("pod", pod.Name), toResourceName("svc", svc.Name), true,
					map[string]string{"dir": "back"})
			}
		}
	}

	return graph.String()
}

func getBinDir() (string, error) {
	s, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(s), nil
}

func normalizeResource(resource string) (string, error) {
	for k, v := range normalizedNames {
		if k == strings.ToLower(resource) {
			return k, nil
		}
		if v == strings.ToLower(resource) {
			return k, nil
		}
	}
	return "", fmt.Errorf("Failed to find normalized resource name for %s", resource)
}

func toImagePath(resource string) string {
	return filepath.Join(dir, "icons", resource+imageSuffix)
}

func escapeName(name string) string {
	return strings.NewReplacer(".", "_", "-", "_").Replace(name)
}

func toClusterName(namespace string) string {
	return clusterPrefix + escapeName(namespace)
}

func toResourceName(resType, name string) string {
	return resType + "_" + escapeName(name)
}

func toRankName(rank int) string {
	return fmt.Sprintf("%s%d", rankPrefix, rank)
}

func toRankDummyNodeName(rank int) string {
	return fmt.Sprintf("%d", rank)
}

func toClusterLabel(namespace string) string {
	return toResourceLabel("ns", namespace)
}

func toResourceLabel(resType, name string) string {
	return fmt.Sprintf("<<TABLE BORDER=\"0\"><TR><TD><IMG SRC=\"%s\" /></TD></TR><TR><TD>%s</TD></TR></TABLE>>", toImagePath(resType), name)
}
