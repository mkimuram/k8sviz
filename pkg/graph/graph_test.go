// SPDX-FileCopyrightText: 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/mkimuram/k8sviz/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	testns       = "testns"
	dir          = "/testdir"
	iconsPath    = "/testdir/icons"
	goldenDir    = "testdata"
	goldenSuffix = ".golden"
	// if -update flag is specified on test run, golden file for the test will be updated
	// Please run:
	// $ go test github.com/mkimuram/k8sviz/pkg/graph -update=true
	update = flag.Bool("update", false, "update the golden files")

	testRes1 = []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs1-pod1",
			Labels:          map[string]string{"app": "rs1"},
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Replicaset", Name: "rs1"}}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs1-pod2",
			Labels:          map[string]string{"app": "rs1"},
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Replicaset", Name: "rs1"}}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs1-pod3",
			Labels:          map[string]string{"app": "rs1"},
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Replicaset", Name: "rs1"}}}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "svc1"},
			Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "rs1"}}},
		&appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs1",
			Labels:          map[string]string{"app": "rs1"},
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: "deploy1"}}}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "deploy1",
			Labels: map[string]string{"app": "rs1"}}},
		&autov1.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "hpa1"},
			Spec: autov1.HorizontalPodAutoscalerSpec{ScaleTargetRef: autov1.CrossVersionObjectReference{Kind: "Deployment", Name: "deploy1", APIVersion: "apps/v1"}}},
		&netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "ing1"},
			Spec: netv1.IngressSpec{Rules: []netv1.IngressRule{
				{IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{
								Path: "/",
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: "svc1",
									},
								},
							},
						},
					},
				}}}}},
	}
	testRes2 = []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1-pod1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Statefulset", Name: "sts1"}}},
			Spec: corev1.PodSpec{Volumes: []corev1.Volume{
				{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "sts1-pvc1",
						},
					},
				},
			}},
		},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1-pod2",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Statefulset", Name: "sts1"}}},
			Spec: corev1.PodSpec{Volumes: []corev1.Volume{
				{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "sts1-pvc2",
						},
					},
				},
			}},
		},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1-pod3",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Statefulset", Name: "sts1"}}},
			Spec: corev1.PodSpec{Volumes: []corev1.Volume{
				{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "sts1-pvc3",
						},
					},
				},
			}},
		},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1-pvc1"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1-pvc2"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1-pvc3"}},
		&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1"}},
	}
	testRes3 = []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "ds1-pod1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "DaemonSet", Name: "ds1"}}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "ds1-pod2",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "DaemonSet", Name: "ds1"}}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "job1-pod1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "Job", Name: "job1"}}}},
		&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "ds1"}},
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "job1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1beta1", Kind: "CronJob", Name: "cronjob1"}}}},
		&batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "cronjob1"}},
	}
)

func prepTestGraph(t *testing.T, objs ...runtime.Object) *Graph {
	cs := fake.NewSimpleClientset(objs...)
	res, err := resources.NewResources(cs, testns)
	if err != nil {
		t.Fatalf("NewResources failed: %v", err)
	}
	return NewGraph(res, dir, iconsPath)
}

func getGoldenFilePath(name string) string {
	return filepath.Join(goldenDir, name+goldenSuffix)
}

func expectedFromGoldenFile(name string) (string, error) {
	content, err := ioutil.ReadFile(getGoldenFilePath(name))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func updateGoldenFile(t *testing.T, name, content string) error {
	if !*update {
		return nil
	}

	goldenFile := getGoldenFilePath(name)
	t.Logf("Updating goldenfile %q\n", goldenFile)
	return ioutil.WriteFile(goldenFile, []byte(content), 0644)
}

func TestGenerateCommon(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "common part of graph for ns=testns and dir=/testdir",
			expected: "common",
		},
	}

	g := prepTestGraph(t)
	for _, tc := range testCases {
		expected, err := expectedFromGoldenFile(tc.expected)
		if err != nil {
			t.Fatalf("[%s] failed to get expected from golden file %s: %v", tc.name, tc.expected, err)
		}

		g.generateCommon()
		dot := g.toDot()

		// Update golden file if -update flag is specified for this test run
		err = updateGoldenFile(t, tc.expected, dot)
		if err != nil {
			t.Fatalf("[%s] failed to update golden file %s: %v", tc.name, tc.expected, err)
		}

		if expected != dot {
			t.Fatalf("[%s] generateCommon doesn't return expected, diff: %v", tc.name, diff.LineDiff(expected, dot))
		}
	}
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name     string
		res      []runtime.Object
		expected string
	}{
		{
			name:     "Generate whole graph for ns=testns and dir=/testdir with testRes1",
			res:      testRes1,
			expected: "generate_res1",
		},
		{
			name:     "Generate whole graph for ns=testns and dir=/testdir with testRes2",
			res:      testRes2,
			expected: "generate_res2",
		},
		{
			name:     "Generate whole graph for ns=testns and dir=/testdir with testRes3",
			res:      testRes3,
			expected: "generate_res3",
		},
	}

	for _, tc := range testCases {
		g := prepTestGraph(t, tc.res...)
		expected, err := expectedFromGoldenFile(tc.expected)
		if err != nil {
			t.Fatalf("[%s] failed to get expected from golden file %s: %v", tc.name, tc.expected, err)
		}

		g.generate()
		dot := g.toDot()

		// Update golden file if -update flag is specified for this test run
		err = updateGoldenFile(t, tc.expected, dot)
		if err != nil {
			t.Fatalf("[%s] failed to update golden file %s: %v", tc.name, tc.expected, err)
		}

		if expected != dot {
			t.Fatalf("[%s] generate doesn't return expected, diff: %v", tc.name, diff.LineDiff(expected, dot))
		}
	}
}
