// SPDX-FileCopyrightText: 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/mkimuram/k8sviz/pkg/resources"
)

// Graph represents a graph of k8s resources
type Graph struct {
	dir  string
	res  *resources.Resources
	gviz *gographviz.Graph
}

// NewGraph returns a Graph of k8s resources
func NewGraph(res *resources.Resources, dir string) *Graph {
	g := &Graph{res: res, dir: dir, gviz: gographviz.NewGraph()}
	g.generate()

	return g
}

// WriteDotFile writes the graph to outFile with dot format
func (g *Graph) WriteDotFile(outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(g.toDot()); err != nil {
		return err
	}

	return nil
}

// PlotDotFile plots the graph to outFile with outType format
func (g *Graph) PlotDotFile(outFile, outType string) error {
	cmd := exec.Command("dot", "-T"+outType, "-o", outFile)
	cmd.Stdin = strings.NewReader(g.toDot())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// toDot returns a string representation of the graph with dot format
func (g *Graph) toDot() string {
	return g.gviz.String()
}

// generate generates the graph of the k8s resources
func (g *Graph) generate() {
	// generate common part of graph
	g.generateCommon()

	// Put resources as Nodes in each rank of subgraph
	g.generateNodes()

	// Connect resources
	g.generateEdges()
}

// generateCommon generates the common part of the graph
func (g *Graph) generateCommon() {
	// Create digraph for namespace.
	// ```
	// digraph G {
	//   rankdir=TD;
	//   label=<<TABLE BORDER="0"><TR><TD><IMG SRC="/icons/ns-128.png" /></TD></TR><TR><TD>ns1</TD></TR></TABLE>>;
	//   labeljust=l;
	//   style=dotted;
	// ```
	g.gviz.SetDir(true)
	g.gviz.SetName("G")
	g.gviz.AddAttr("G", "rankdir", "TD")
	g.gviz.AddSubGraph("G", g.clusterName(),
		map[string]string{"label": g.clusterLabel(), "labeljust": "l", "style": "dotted"})

	// Create subgraphs for resources to group by rank (repeats #ResourceTypes)
	// ```
	// subgraph rank_0 {
	// rank=same;
	// style=invis;
	// 0 [ height=0, margin=0, style=invis, width=0 ];
	// }
	// ;
	//
	// subgraph rank_1 {
	// rank=same;
	// style=invis;
	// 1 [ height=0, margin=0, style=invis, width=0 ];
	// }
	// ;
	// ```
	for r := 0; r < len(resources.ResourceTypes); r++ {
		g.gviz.AddSubGraph(g.clusterName(), g.rankName(r),
			map[string]string{"rank": "same", "style": "invis"})
		// Put dummy invisible node to order ranks
		g.gviz.AddNode(g.rankName(r), g.rankDummyNodeName(r),
			map[string]string{"style": "invis", "height": "0", "width": "0", "margin": "0"})
	}

	// Order ranks (repeats #ResourceTypes)
	// This will make the layout consistent.
	// ```
	// 0->1[ style=invis ];
	// 1->2[ style=invis ];
	// ```
	for r := 0; r < len(resources.ResourceTypes)-1; r++ {
		// Connect rth node and r+1th dummy node with invisible edge
		g.gviz.AddEdge(g.rankDummyNodeName(r), g.rankDummyNodeName(r+1), true,
			map[string]string{"style": "invis"})
	}
}

// generateNodes generates the nodes of the graph
// K8s resources are represented as graph nodes in k8sviz.
func (g *Graph) generateNodes() {
	// Create graphviz nodes for k8s resources like below.
	// ```
	// pod_my_pod [ label=<<TABLE BORDER="0"><TR><TD><IMG SRC="/icons/pod-128.png" /></TD></TR><TR><TD>my-pod</TD></TR></TABLE>>, penwidth=0 ];
	// ```
	// Each resource is created in the subgraph of the rank for its resource types,
	// so that the same resource types are placed in the same rank.
	for r, rankRes := range resources.ResourceTypes {
		for _, resType := range strings.Fields(rankRes) {
			for _, name := range g.res.GetResourceNames(resType) {
				g.gviz.AddNode(g.rankName(r), g.resourceName(resType, name),
					map[string]string{"label": g.resourceLabel(resType, name), "penwidth": "0"})
			}
		}
	}
}

// generateEdges generates the edges of the graph
// Relations between k8s resources are represented as graph edges in k8sviz.
func (g *Graph) generateEdges() {
	// Owner reference for pod
	g.genPodOwnerRef()

	// Owner reference for rs
	g.genRsOwnerRef()

	// pvc and pod
	g.genPvcPodRef()

	// svc and pod
	g.genSvcPodRef()

	// ingress and svc
	g.genIngSvcRef()
}

// genPodOwnerRef generates the edges of OwnerReferences from Pod
func (g *Graph) genPodOwnerRef() {
	// Add edge if below matches:
	//   - v1.Pod.metadata.ownerReferences.
	//     - kind
	//     - name
	//   - {kind}.metadata.{name}
	// ```
	// rs_my_replicaset->pod_my_pod [ style=dashed ];
	// ```
	for _, pod := range g.res.Pods.Items {
		for _, ref := range pod.GetOwnerReferences() {
			ownerKind, err := resources.NormalizeResource(ref.Kind)
			if err != nil {
				// Skip resource that isn't available for this tool, like CRD
				continue
			}
			if !g.res.HasResource(ownerKind, ref.Name) {
				fmt.Fprintf(os.Stderr, "%s %s not found as a owner refernce for po %s\n", ownerKind, ref.Name, pod.Name)
				continue
			}
			g.gviz.AddEdge(g.resourceName(ownerKind, ref.Name), g.resourceName("pod", pod.Name), true,
				map[string]string{"style": "dashed"})
		}
	}
}

// genRsOwnerRef generates the edges of OwnerReferences from RS
func (g *Graph) genRsOwnerRef() {
	// Add edge if below matches:
	//   - apps/v1.ReplicaSet.metadata.ownerReferences.
	//     - kind
	//     - name
	//   - {kind}.metadata.{name}
	// ```
	// deploy_my_deployment->rs_my_replicaset[ style=dashed ];
	// ```
	for _, rs := range g.res.Rss.Items {
		for _, ref := range rs.GetOwnerReferences() {
			ownerKind, err := resources.NormalizeResource(ref.Kind)
			if err != nil {
				// Skip resource that isn't available for this tool, like CRD
				continue
			}
			if !g.res.HasResource(ownerKind, ref.Name) {
				fmt.Fprintf(os.Stderr, "%s %s not found as a owner refernce for rs %s\n", ownerKind, ref.Name, rs.Name)
				continue
			}

			g.gviz.AddEdge(g.resourceName(ownerKind, ref.Name), g.resourceName("rs", rs.Name), true,
				map[string]string{"style": "dashed"})
		}
	}
}

// genPvcPodRef generates the edges of PVC to Pod reference
func (g *Graph) genPvcPodRef() {
	// Add edge if below matches:
	//   - v1.Pod.spec.volumes[].persistentVolumeClaim.claimName
	//   - v1.PersistentVolumeClaim.metadata.name
	// ```
	// pod_my_pod->pvc_my_persistentvolumeclaim[ dir=none ];
	// ```
	for _, pod := range g.res.Pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.VolumeSource.PersistentVolumeClaim != nil {
				if !g.res.HasResource("pvc", vol.VolumeSource.PersistentVolumeClaim.ClaimName) {
					fmt.Fprintf(os.Stderr, "pvc %s not found as a volume for pod %s\n", vol.VolumeSource.PersistentVolumeClaim.ClaimName, pod.Name)
					continue
				}

				g.gviz.AddEdge(g.resourceName("pod", pod.Name), g.resourceName("pvc", vol.VolumeSource.PersistentVolumeClaim.ClaimName), true,
					map[string]string{"dir": "none"})
			}
		}
	}
}

// genSvcPodRef generates the edges of Service to Pod reference
func (g *Graph) genSvcPodRef() {
	// Add edge if below matches:
	//   - v1.Service.spec.selector
	//   - v1.Pod.metadata.labels
	// ```
	// pod_my_pod->svc_my_service[ dir=back ];
	// ```
	for _, svc := range g.res.Svcs.Items {
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		// Check if pod has all labels specified in svc.Spec.Selector
		for _, pod := range g.res.Pods.Items {
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
				g.gviz.AddEdge(g.resourceName("pod", pod.Name), g.resourceName("svc", svc.Name), true,
					map[string]string{"dir": "back"})
			}
		}
	}
}

// genIngSvcRef generates the edges of Ingress to Service reference
func (g *Graph) genIngSvcRef() {
	// Add edge if below matches:
	//   - networking.k8s.io/v1beta1.Ingress.spec.rules.path[].backend.serviceName
	//   - v1.Service.metadata.name
	// ```
	// svc_my_service->ing_my_ingress[ dir=back ];
	// ```
	for _, ing := range g.res.Ingresses.Items {
		for _, rule := range ing.Spec.Rules {
			for _, path := range rule.IngressRuleValue.HTTP.Paths {
				if !g.res.HasResource("svc", path.Backend.ServiceName) {
					fmt.Fprintf(os.Stderr, "svc %s not found for ingress %s\n", path.Backend.ServiceName, ing.Name)
					continue
				}

				g.gviz.AddEdge(g.resourceName("svc", path.Backend.ServiceName), g.resourceName("ing", ing.Name), true, map[string]string{"dir": "back"})
			}
		}
	}
}

// imagePath returns the path to the image file
// path is {dir}/icons/{resource}-128.png
// ex) /icons/pod-128.png
func (g *Graph) imagePath(resource string) string {
	return filepath.Join(g.dir, "icons", resource+imageSuffix)
}

// clusterLabel returns the resource label for namespace
// ex)
//   <<TABLE BORDER="0"><TR><IMG SRC="/icons/ns-128.png" /></TR><TR><TD>my-namespace</TD></TR></TABLE>>
func (g *Graph) clusterLabel() string {
	return g.resourceLabel("ns", g.res.Namespace)
}

// resourceLabel returns the resource label for a resource
// ex)
//   <<TABLE BORDER="0"><TR><IMG SRC="/icons/pod-128.png" /></TR><TR><TD>my-pod</TD></TR></TABLE>>
func (g *Graph) resourceLabel(resType, name string) string {
	return fmt.Sprintf("<<TABLE BORDER=\"0\"><TR><TD><IMG SRC=\"%s\" /></TD></TR><TR><TD>%s</TD></TR></TABLE>>", g.imagePath(resType), name)
}

// clusterName returns name of the graphviz cluster
// It is named base on namespace.
// ex) cluster_my_namespace
func (g *Graph) clusterName() string {
	return clusterPrefix + g.escapeName(g.res.Namespace)
}

// escapeName returns the escaped name to be handled with graphviz
// It replaces "." and "-" with "_".
// ex) my_namespace
func (g *Graph) escapeName(name string) string {
	return strings.NewReplacer(".", "_", "-", "_").Replace(name)
}

// resourceName returns the escaped name of the resource
// It espaces the resource name and add resType as a prefix.
// ex) pod_my_pod
func (g *Graph) resourceName(resType, name string) string {
	return resType + "_" + g.escapeName(name)
}

// rankName returns the name of the dummy rank
// ex) rank_1
func (g *Graph) rankName(rank int) string {
	return fmt.Sprintf("%s%d", rankPrefix, rank)
}

// rankDummyNodeName returns the node name of the dummy rank
// ex) 1
func (g *Graph) rankDummyNodeName(rank int) string {
	return fmt.Sprintf("%d", rank)
}
