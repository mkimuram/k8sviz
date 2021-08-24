// SPDX-FileCopyrightText: 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	zero      = int32(0)
	one       = int32(1)
	testns    = "testns"
	nontestns = "nontestns"
	testRes1  = []runtime.Object{
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "pod1"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: nontestns, Name: "pod2"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "svc1"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "svc2"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "pvc1"}},
		&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "sts1"}},
		&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "ds1"}},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs1"},
			Spec:       appsv1.ReplicaSetSpec{Replicas: &one},
			Status:     appsv1.ReplicaSetStatus{Replicas: int32(1)},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs2"},
			Spec:       appsv1.ReplicaSetSpec{Replicas: &zero},
			Status:     appsv1.ReplicaSetStatus{Replicas: int32(1)},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs3"},
			Spec:       appsv1.ReplicaSetSpec{Replicas: &one},
			Status:     appsv1.ReplicaSetStatus{Replicas: int32(0)},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "rs4"},
			Spec:       appsv1.ReplicaSetSpec{Replicas: &zero},
			Status:     appsv1.ReplicaSetStatus{Replicas: int32(0)},
		},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "deploy1"}},
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "job1"}},
		&batchv1beta1.CronJob{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "cronjob1"}},
		&netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "ing1"}},
		&autov1.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: testns, Name: "hpa1"}},
	}
)

func TestGetResourceNames(t *testing.T) {
	testCases := []struct {
		name      string
		resources []runtime.Object
		kind      string
		expected  []string
	}{
		{
			name:      "No resources and kind:pod is specified",
			resources: []runtime.Object{},
			kind:      "pod",
			expected:  []string{},
		},
		{
			name:      "Un-known kind is specified",
			resources: testRes1,
			kind:      "unknown",
			expected:  []string{},
		},
		{
			name:      "pod1 in testns and kind:pod is specified",
			resources: testRes1,
			kind:      "pod",
			expected:  []string{"pod1"},
		},
		{
			name:      "svc1 and svc2 in testns and kind:svc is specified",
			resources: testRes1,
			kind:      "svc",
			expected:  []string{"svc1", "svc2"},
		},
		{
			name:      "pvc1 in testns and kind:pvc is specified",
			resources: testRes1,
			kind:      "pvc",
			expected:  []string{"pvc1"},
		},
		{
			name:      "sts1 in testns and kind:sts is specified",
			resources: testRes1,
			kind:      "sts",
			expected:  []string{"sts1"},
		},
		{
			name:      "ds1 in testns and kind:ds is specified",
			resources: testRes1,
			kind:      "ds",
			expected:  []string{"ds1"},
		},
		{
			name:      "rs1, rs2, rs3 and rs4(with no replica) in testns and kind:rs is specified",
			resources: testRes1,
			kind:      "rs",
			expected:  []string{"rs1", "rs2", "rs3"},
		},
		{
			name:      "deploy1 in testns and kind:deploy is specified",
			resources: testRes1,
			kind:      "deploy",
			expected:  []string{"deploy1"},
		},
		{
			name:      "job1 in testns and kind:job is specified",
			resources: testRes1,
			kind:      "job",
			expected:  []string{"job1"},
		},
		{
			name:      "cronjob1 in testns and kind:cronjob is specified",
			resources: testRes1,
			kind:      "cronjob",
			expected:  []string{"cronjob1"},
		},
		{
			name:      "ing1 in testns and kind:ingress is specified",
			resources: testRes1,
			kind:      "ing",
			expected:  []string{"ing1"},
		},
		{
			name:      "hpa1 in testns and kind:hpa is specified",
			resources: testRes1,
			kind:      "hpa",
			expected:  []string{"hpa1"},
		},
	}

	for _, tc := range testCases {
		cs := fake.NewSimpleClientset(tc.resources...)
		res, err := NewResources(cs, testns)
		if err != nil {
			t.Fatalf("NewResources failed: %v", err)
		}

		resNames := res.GetResourceNames(tc.kind)

		// Check if resNames and tc.expcted are exact the same
		if len(tc.expected) != len(resNames) {
			t.Fatalf("[%s] GetResourceNames doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, resNames)
		}

		expectedMap := map[string]bool{}
		for _, expected := range tc.expected {
			expectedMap[expected] = true
		}
		for _, res := range resNames {
			_, ok := expectedMap[res]
			if !ok {
				t.Fatalf("[%s] GetResourceNames doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, resNames)
			}
		}
	}
}

func TestHasResource(t *testing.T) {
	testCases := []struct {
		name         string
		resources    []runtime.Object
		kind         string
		resourceName string
		expected     bool
	}{
		{
			name:         "No resources and kind:pod is specified",
			resources:    []runtime.Object{},
			kind:         "pod",
			resourceName: "pod1",
			expected:     false,
		},
		{
			name:         "One resource per kind in testns and un-known/pod1 is specified",
			resources:    testRes1,
			kind:         "unknown",
			resourceName: "pod1",
			expected:     false,
		},
		{
			name:         "pod1 in testns and pod/pod1 is specified",
			resources:    testRes1,
			kind:         "pod",
			resourceName: "pod1",
			expected:     true,
		},
		{
			name:         "pod2 in nontestns and pod/pod2 is specified",
			resources:    testRes1,
			kind:         "pod",
			resourceName: "pod2",
			expected:     false,
		},
		{
			name:         "svc/svc1 in testns and pod/svc1 is specified",
			resources:    testRes1,
			kind:         "pod",
			resourceName: "svc1",
			expected:     false,
		},
		{
			name:         "svc1 in testns and svc/svc1 is specified",
			resources:    testRes1,
			kind:         "svc",
			resourceName: "svc1",
			expected:     true,
		},
		{
			name:         "svc2 in testns and svc/svc2 is specified",
			resources:    testRes1,
			kind:         "svc",
			resourceName: "svc2",
			expected:     true,
		},
		{
			name:         "pvc1 in testns and pvc/pvc1 is specified",
			resources:    testRes1,
			kind:         "pvc",
			resourceName: "pvc1",
			expected:     true,
		},
		{
			name:         "sts1 in testns and sts/sts1 is specified",
			resources:    testRes1,
			kind:         "sts",
			resourceName: "sts1",
			expected:     true,
		},
		{
			name:         "ds1 in testns and ds/ds1 is specified",
			resources:    testRes1,
			kind:         "ds",
			resourceName: "ds1",
			expected:     true,
		},
		{
			name:         "rs1 in testns and rs/rs1 is specified",
			resources:    testRes1,
			kind:         "rs",
			resourceName: "rs1",
			expected:     true,
		},
		{
			name:         "deploy1 in testns and deploy/deploy1 is specified",
			resources:    testRes1,
			kind:         "deploy",
			resourceName: "deploy1",
			expected:     true,
		},
		{
			name:         "job1 in testns and job/job1 is specified",
			resources:    testRes1,
			kind:         "job",
			resourceName: "job1",
			expected:     true,
		},
		{
			name:         "cronjob1 in testns and cronjob/cronjob1 is specified",
			resources:    testRes1,
			kind:         "cronjob",
			resourceName: "cronjob1",
			expected:     true,
		},
		{
			name:         "ing1 in testns and ing/ing1 is specified",
			resources:    testRes1,
			kind:         "ing",
			resourceName: "ing1",
			expected:     true,
		},
		{
			name:         "hpa1 in testns and kind:hpa is specified",
			resources:    testRes1,
			kind:         "hpa",
			resourceName: "hpa1",
			expected:     true,
		},
	}

	for _, tc := range testCases {
		cs := fake.NewSimpleClientset(tc.resources...)
		res, err := NewResources(cs, testns)
		if err != nil {
			t.Fatalf("NewResources failed: %v", err)
		}

		has := res.HasResource(tc.kind, tc.resourceName)

		// Check if has and tc.expcted are the same
		if tc.expected != has {
			t.Fatalf("[%s] HasResource doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, has)
		}
	}
}

func TestNormalizeResource(t *testing.T) {
	testCases := []struct {
		name      string
		kind      string
		expected  string
		expectErr bool
	}{
		{
			name:      "Should return error for unknown kind",
			kind:      "unknown",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Should return ns for ns",
			kind:      "ns",
			expected:  "ns",
			expectErr: false,
		},
		{
			name:      "Should return ns for namespace",
			kind:      "namespace",
			expected:  "ns",
			expectErr: false,
		},
		{
			name:      "Should return ns for NameSpace",
			kind:      "NameSpace",
			expected:  "ns",
			expectErr: false,
		},
		{
			name:      "Should return ns for NS",
			kind:      "NS",
			expected:  "ns",
			expectErr: false,
		},
		{
			name:      "Should return svc for svc",
			kind:      "svc",
			expected:  "svc",
			expectErr: false,
		},
		{
			name:      "Should return svc for service",
			kind:      "service",
			expected:  "svc",
			expectErr: false,
		},
		{
			name:      "Should return pvc for persistentvolumeclaim",
			kind:      "persistentvolumeclaim",
			expected:  "pvc",
			expectErr: false,
		},
		{
			name:      "Should return pvc for pvc",
			kind:      "pvc",
			expected:  "pvc",
			expectErr: false,
		},
		{
			name:      "Should return pod for po",
			kind:      "po",
			expected:  "pod",
			expectErr: false,
		},
		{
			name:      "Should return pod for po",
			kind:      "pod",
			expected:  "pod",
			expectErr: false,
		},
		{
			name:      "Should return sts for statefulset",
			kind:      "statefulset",
			expected:  "sts",
			expectErr: false,
		},
		{
			name:      "Should return sts for sts",
			kind:      "sts",
			expected:  "sts",
			expectErr: false,
		},
		{
			name:      "Should return ds for daemonset",
			kind:      "daemonset",
			expected:  "ds",
			expectErr: false,
		},
		{
			name:      "Should return ds for ds",
			kind:      "ds",
			expected:  "ds",
			expectErr: false,
		},
		{
			name:      "Should return rs for replicaset",
			kind:      "replicaset",
			expected:  "rs",
			expectErr: false,
		},
		{
			name:      "Should return rs for rs",
			kind:      "rs",
			expected:  "rs",
			expectErr: false,
		},
		{
			name:      "Should return deploy for deployment",
			kind:      "deployment",
			expected:  "deploy",
			expectErr: false,
		},
		{
			name:      "Should return deploy for deploy",
			kind:      "deploy",
			expected:  "deploy",
			expectErr: false,
		},
		{
			name:      "Should return job for job",
			kind:      "job",
			expected:  "job",
			expectErr: false,
		},
		{
			name:      "Should return ing for ingress",
			kind:      "ingress",
			expected:  "ing",
			expectErr: false,
		},
		{
			name:      "Should return ing for ing",
			kind:      "ing",
			expected:  "ing",
			expectErr: false,
		},
		{
			name:      "Should return hpa for hpa",
			kind:      "hpa",
			expected:  "hpa",
			expectErr: false,
		},
		{
			name:      "Should return hpa for HorizontalPodAutoscaler",
			kind:      "HorizontalPodAutoscaler",
			expected:  "hpa",
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		normalized, err := NormalizeResource(tc.kind)
		// Check if tc.expcted and normalized are the same
		if tc.expected != normalized {
			t.Fatalf("[%s] NormalizeResource doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, normalized)
		}
		// Check if NormalizeResource returns error if tc.expctErr is true
		if tc.expectErr && err == nil {
			t.Fatalf("[%s] NormalizeResource expects error, but returned no error", tc.name)
		}
		// Check if NormalizeResource returns no error if tc.expctErr is false
		if !tc.expectErr && err != nil {
			t.Fatalf("[%s] NormalizeResource expects no error, but returned error %v", tc.name, err)
		}
	}
}
