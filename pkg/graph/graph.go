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

type Graph struct {
	dir  string
	res  *resources.Resources
	gviz *gographviz.Graph
}

func NewGraph(res *resources.Resources, dir string) *Graph {
	g := &Graph{res: res, dir: dir, gviz: gographviz.NewGraph()}
	g.generate()

	return g
}

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

func (g *Graph) toDot() string {
	return g.gviz.String()
}

func (g *Graph) generate() {
	// generate common part of graph
	g.generateCommon()

	// Put resources as Nodes in each rank of subgraph
	g.generateNodes()

	// Connect resources
	g.generateEdges()
}

func (g *Graph) generateCommon() {
	g.gviz.SetDir(true)
	g.gviz.SetName("G")
	g.gviz.AddAttr("G", "rankdir", "TD")
	g.gviz.AddSubGraph("G", g.clusterName(),
		map[string]string{"label": g.clusterLabel(), "labeljust": "l", "style": "dotted"})

	// Create subgraphs for resources to group by rank
	for r := 0; r < len(resources.ResourceTypes); r++ {
		g.gviz.AddSubGraph(g.clusterName(), g.rankName(r),
			map[string]string{"rank": "same", "style": "invis"})
		// Put dummy invisible node to order ranks
		g.gviz.AddNode(g.rankName(r), g.rankDummyNodeName(r),
			map[string]string{"style": "invis", "height": "0", "width": "0", "margin": "0"})
	}

	// Order ranks
	for r := 0; r < len(resources.ResourceTypes)-1; r++ {
		// Connect rth node and r+1th dummy node with invisible edge
		g.gviz.AddEdge(g.rankDummyNodeName(r), g.rankDummyNodeName(r+1), true,
			map[string]string{"style": "invis"})
	}
}

func (g *Graph) generateNodes() {
	for r, rankRes := range resources.ResourceTypes {
		for _, resType := range strings.Fields(rankRes) {
			for _, name := range g.res.GetResourceNames(resType) {
				g.gviz.AddNode(g.rankName(r), g.resourceName(resType, name),
					map[string]string{"label": g.resourceLabel(resType, name), "penwidth": "0"})
			}
		}
	}
}

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

func (g *Graph) genPodOwnerRef() {
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

func (g *Graph) genRsOwnerRef() {
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

func (g *Graph) genPvcPodRef() {
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

func (g *Graph) genSvcPodRef() {
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

func (g *Graph) genIngSvcRef() {
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

func (g *Graph) imagePath(resource string) string {
	return filepath.Join(g.dir, "icons", resource+imageSuffix)
}

func (g *Graph) clusterLabel() string {
	return g.resourceLabel("ns", g.res.Namespace)
}

func (g *Graph) resourceLabel(resType, name string) string {
	return fmt.Sprintf("<<TABLE BORDER=\"0\"><TR><TD><IMG SRC=\"%s\" /></TD></TR><TR><TD>%s</TD></TR></TABLE>>", g.imagePath(resType), name)
}

func (g *Graph) clusterName() string {
	return clusterPrefix + g.escapeName(g.res.Namespace)
}

func (g *Graph) escapeName(name string) string {
	return strings.NewReplacer(".", "_", "-", "_").Replace(name)
}

func (g *Graph) resourceName(resType, name string) string {
	return resType + "_" + g.escapeName(name)
}

func (g *Graph) rankName(rank int) string {
	return fmt.Sprintf("%s%d", rankPrefix, rank)
}

func (g *Graph) rankDummyNodeName(rank int) string {
	return fmt.Sprintf("%d", rank)
}
