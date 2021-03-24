// SPDX-FileCopyrightText: 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// ResourceTypes represents the set of resource types.
	// Resouces are grouped by the same level of abstraction.
	ResourceTypes   = []string{"hpa", "deploy job", "sts ds rs", "pod", "pvc", "svc", "ing"}
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
		"ing":    "ingress",
		"hpa":    "horizontalpodautoscaler",
	}
)

// Resources represents the k8s resources
type Resources struct {
	clientset kubernetes.Interface
	Namespace string

	Svcs      *corev1.ServiceList
	Pvcs      *corev1.PersistentVolumeClaimList
	Pods      *corev1.PodList
	Stss      *appsv1.StatefulSetList
	Dss       *appsv1.DaemonSetList
	Rss       *appsv1.ReplicaSetList
	Deploys   *appsv1.DeploymentList
	Jobs      *batchv1.JobList
	Ingresses *v1beta1.IngressList
	Hpas      *autov1.HorizontalPodAutoscalerList
}

// NewResources resturns Resources for the namespace
func NewResources(clientset kubernetes.Interface, namespace string) (*Resources, error) {
	var err error
	res := &Resources{clientset: clientset, Namespace: namespace}

	// service
	res.Svcs, err = clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get services in namespace %q: %v", namespace, err)
	}

	// persistentvolumeclaim
	res.Pvcs, err = clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get persistentVolumeClaims in namespace %q: %v", namespace, err)
	}

	// pod
	res.Pods, err = clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods in namespace %q: %v", namespace, err)
	}

	// statefulset
	res.Stss, err = clientset.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulsets in namespace %q: %v", namespace, err)
	}

	// daemonset
	res.Dss, err = clientset.AppsV1().DaemonSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonsets in namespace %q: %v", namespace, err)
	}

	// replicaset
	res.Rss, err = clientset.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get replicasets in namespace %q: %v", namespace, err)
	}

	// deployment
	res.Deploys, err = clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments in namespace %q: %v", namespace, err)
	}

	// job
	res.Jobs, err = clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs in namespace %q: %v", namespace, err)
	}

	// ingress
	res.Ingresses, err = clientset.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingresses in namespace %q: %v", namespace, err)
	}

	// hpas
	res.Hpas, err = clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get hpas in namespace %q: %v", namespace, err)
	}

	return res, nil
}

// GetResourceNames returns the resource names of the kind
func (r *Resources) GetResourceNames(kind string) []string {
	names := []string{}

	switch kind {
	case "svc":
		for _, n := range r.Svcs.Items {
			names = append(names, n.Name)
		}
	case "pvc":
		for _, n := range r.Pvcs.Items {
			names = append(names, n.Name)
		}
	case "pod":
		for _, n := range r.Pods.Items {
			names = append(names, n.Name)
		}
	case "sts":
		for _, n := range r.Stss.Items {
			names = append(names, n.Name)
		}
	case "ds":
		for _, n := range r.Dss.Items {
			names = append(names, n.Name)
		}
	case "rs":
		for _, n := range r.Rss.Items {
			names = append(names, n.Name)
		}
	case "deploy":
		for _, n := range r.Deploys.Items {
			names = append(names, n.Name)
		}
	case "job":
		for _, n := range r.Jobs.Items {
			names = append(names, n.Name)
		}
	case "ing":
		for _, n := range r.Ingresses.Items {
			names = append(names, n.Name)
		}
	case "hpa":
		for _, n := range r.Hpas.Items {
			names = append(names, n.Name)
		}
	}

	return names
}

// HasResource check if Resources has k8s resource with the kind and the name
func (r *Resources) HasResource(kind, name string) bool {
	for _, resName := range r.GetResourceNames(kind) {
		if resName == name {
			return true
		}
	}
	return false
}

// NormalizeResource resturns normalized name of the resource.
// It returns error if it fails to normalize the resource name.
// key of normalizedNames map is used as the normalized name.
func NormalizeResource(resource string) (string, error) {
	for k, v := range normalizedNames {
		if k == strings.ToLower(resource) {
			return k, nil
		}
		if v == strings.ToLower(resource) {
			return k, nil
		}
	}
	return "", fmt.Errorf("failed to find normalized resource name for %s", resource)
}
