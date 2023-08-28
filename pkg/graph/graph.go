// SPDX-FileCopyrightText: 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/mkimuram/k8sviz/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Graph represents a graph of k8s resources
type Graph struct {
	dir       string
	iconsPath string
	res       *resources.Resources
	gviz      *gographviz.Graph
}

// NewGraph returns a Graph of k8s resources
func NewGraph(res *resources.Resources, dir, iconsPath string) *Graph {
	g := &Graph{res: res, dir: dir, iconsPath: iconsPath, gviz: gographviz.NewGraph()}
	g.generate()

	return g
}

// WriteDotFile writes the graph to outFile with dot format
func (g *Graph) WriteDotFile(outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(g.toDot()); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return fmt.Errorf("failed to close file after write failure: %v, %v", closeErr, err)
		}
		return err
	}

	return f.Close()
}

// PlotDotFile plots the graph to outFile with outType format
func (g *Graph) PlotDotFile(outFile, outType string) error {
	var cmd *exec.Cmd

	// To avoid CWE-78, passing static argument to exec.Command
	switch outType {
	case "ps":
		cmd = exec.Command("dot", "-Tps")
	case "pdf":
		cmd = exec.Command("dot", "-Tpdf")
	case "svg":
		cmd = exec.Command("dot", "-Tsvg")
	case "png":
		cmd = exec.Command("dot", "-Tpng")
	case "gif":
		cmd = exec.Command("dot", "-Tgif")
	case "jpg":
		cmd = exec.Command("dot", "-Tjpg")
	default:
		return fmt.Errorf("format %q is not supported", outType)
	}

	// Call dot command
	var stdout, stderr bytes.Buffer
	cmd.Stdin = strings.NewReader(g.toDot())
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create dot file: stderr: %v, err: %v", stderr.String(), err)
	}

	// Write to outFile
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(stdout.String()); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return fmt.Errorf("failed to close file after write failure: %v, %v", closeErr, err)
		}
		return err
	}

	return f.Close()
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
	err := g.gviz.SetDir(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set to digraph: %v\n", err)
	}
	err = g.gviz.SetName("G")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set to graph name to G: %v\n", err)
	}
	err = g.gviz.AddAttr("G", "rankdir", "TD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set rankdir to TD: %v\n", err)
	}
	err = g.gviz.AddSubGraph("G", g.clusterName(),
		map[string]string{"label": g.clusterLabel(), "labeljust": "l", "style": "dotted"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add subgraph %s to digraph G: %v\n", g.clusterName(), err)
	}

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
		err = g.gviz.AddSubGraph(g.clusterName(), g.rankName(r),
			map[string]string{"rank": "same", "style": "invis"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add subgraph %s to subgraph %s: %v\n", g.rankName(r), g.clusterName(), err)
		}

		// Put dummy invisible node to order ranks
		err = g.gviz.AddNode(g.rankName(r), g.rankDummyNodeName(r),
			map[string]string{"style": "invis", "height": "0", "width": "0", "margin": "0"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add node %s to subgraph %s: %v\n", g.rankDummyNodeName(r), g.rankName(r), err)
		}
	}

	// Order ranks (repeats #ResourceTypes)
	// This will make the layout consistent.
	// ```
	// 0->1[ style=invis ];
	// 1->2[ style=invis ];
	// ```
	for r := 0; r < len(resources.ResourceTypes)-1; r++ {
		// Connect rth node and r+1th dummy node with invisible edge
		err = g.gviz.AddEdge(g.rankDummyNodeName(r), g.rankDummyNodeName(r+1), true,
			map[string]string{"style": "invis"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add edge from %s to %s: %v\n", g.rankDummyNodeName(r), g.rankDummyNodeName(r+1), err)
		}
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
				err := g.gviz.AddNode(g.rankName(r), g.resourceName(resType, name),
					map[string]string{"label": g.resourceLabel(resType, name), "penwidth": "0"})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add node %s to subgraph %s: %v\n", g.resourceName(resType, name), g.rankName(r), err)
				}
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

	// Owner reference for job
	g.genJobOwnerRef()

	// hpa to scale target
	g.genHpaScaleTargetRef()

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
		g.genOwnerRef("pod", &pod)
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
		g.genOwnerRef("rs", &rs)
	}
}

// genJobOwnerRef generates the edges of OwnerReferences from job
func (g *Graph) genJobOwnerRef() {
	// Add edge if below matches:
	//   - batch/v1.Job.metadata.ownerReferences.
	//     - kind
	//     - name
	//   - {kind}.metadata.{name}
	// ```
	// cronjob_my_cronjob->job_my_job[ style=dashed ];
	// ```
	for _, job := range g.res.Jobs.Items {
		g.genOwnerRef("job", &job)
	}
}

// genOwnerRef generates the edges of OwnerReferences for specified obj
func (g *Graph) genOwnerRef(kind string, obj metav1.Object) {
	for _, ref := range obj.GetOwnerReferences() {
		ownerKind, err := resources.NormalizeResource(ref.Kind)
		if err != nil {
			// Skip resource that isn't available for this tool, like CRD
			continue
		}
		if !g.res.HasResource(ownerKind, ref.Name) {
			fmt.Fprintf(os.Stderr, "%s %s not found as a owner refernce for rs %s\n", ownerKind, ref.Name, obj.GetName())
			continue
		}

		err = g.gviz.AddEdge(g.resourceName(ownerKind, ref.Name), g.resourceName(kind, obj.GetName()), true,
			map[string]string{"style": "dashed"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add edge from %s to %s: %v\n", g.resourceName(ownerKind, ref.Name), g.resourceName(kind, obj.GetName()), err)
		}
	}
}

// genHpaScaleTargetRef generates the edges of HPA to deploy reference
func (g *Graph) genHpaScaleTargetRef() {
	// Add edge if below matches:
	//   - k8s.io/api/autoscaling/v1.HorizontalPodAutoscaler.Spec.ScaleTargetRef.
	//     - kind
	//     - name
	//   - {kind}.metadata.{name}
	// ```
	// hpa_my_hpa->deploy_my_deploy[ style=dashed ];
	// ```
	for _, hpa := range g.res.Hpas.Items {
		target := hpa.Spec.ScaleTargetRef
		targetKind, err := resources.NormalizeResource(target.Kind)
		if err != nil {
			// Skip resource that isn't available for this tool, like CRD
			continue
		}
		if !g.res.HasResource(targetKind, target.Name) {
			fmt.Fprintf(os.Stderr, "%s %q is referenced from %q, but not found\n", targetKind, target.Name, hpa.Name)
			continue
		}

		err = g.gviz.AddEdge(g.resourceName("hpa", hpa.Name), g.resourceName(targetKind, target.Name), true, map[string]string{"style": "dashed"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add edge from %s to %s: %v\n", g.resourceName("hpa", hpa.Name), g.resourceName(targetKind, target.Name), err)
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

				err := g.gviz.AddEdge(g.resourceName("pod", pod.Name), g.resourceName("pvc", vol.VolumeSource.PersistentVolumeClaim.ClaimName), true, map[string]string{"dir": "none"})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add edge from %s to %s: %v\n", g.resourceName("pod", pod.Name), g.resourceName("pvc", vol.VolumeSource.PersistentVolumeClaim.ClaimName), err)
				}

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
				err := g.gviz.AddEdge(g.resourceName("pod", pod.Name), g.resourceName("svc", svc.Name), true, map[string]string{"dir": "back"})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add edge from %s to %s: %v\n", g.resourceName("pod", pod.Name), g.resourceName("svc", svc.Name), err)
				}
			}
		}
	}
}

// genIngSvcRef generates the edges of Ingress to Service reference
func (g *Graph) genIngSvcRef() {
	// Add edge if below matches:
	//   - networking.k8s.io/v1.Ingress.spec.rules.HTTP.paths[].backend.service.name
	//   - v1.Service.metadata.name
	// ```
	// svc_my_service->ing_my_ingress[ dir=back ];
	// ```
	for _, ing := range g.res.Ingresses.Items {
		for _, rule := range ing.Spec.Rules {
			for _, path := range rule.IngressRuleValue.HTTP.Paths {
				if !g.res.HasResource("svc", path.Backend.Service.Name) {
					fmt.Fprintf(os.Stderr, "svc %s not found for ingress %s\n", path.Backend.Service.Name, ing.Name)
					continue
				}

				err := g.gviz.AddEdge(g.resourceName("svc", path.Backend.Service.Name), g.resourceName("ing", ing.Name), true, map[string]string{"dir": "back"})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add edge from %s to %s: %v\n", g.resourceName("svc", path.Backend.Service.Name), g.resourceName("ing", ing.Name), err)
				}
			}
		}
	}
}
