// SPDX-FileCopyrightText: 2021 k8sviz authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"testing"
)

func TestImagePath(t *testing.T) {
	testCases := []struct {
		name     string
		kind     string
		expected string
	}{
		{
			name:     "kind:pod is specified",
			kind:     "pod",
			expected: "/testdir/icons/pod-128.png",
		},
		{
			name:     "kind:svc is specified",
			kind:     "svc",
			expected: "/testdir/icons/svc-128.png",
		},
	}

	g := prepTestGraph(t)
	for _, tc := range testCases {
		path := g.imagePath(tc.kind)
		if tc.expected != path {
			t.Fatalf("[%s] imagePath doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, path)
		}
	}
}

func TestClusterLabel(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "For namespace=testns and dir=/testdir",
			expected: "<<TABLE BORDER=\"0\"><TR><TD><IMG SRC=\"/testdir/icons/ns-128.png\" /></TD></TR><TR><TD>testns</TD></TR></TABLE>>",
		},
	}

	g := prepTestGraph(t)
	for _, tc := range testCases {
		label := g.clusterLabel()
		if tc.expected != label {
			t.Fatalf("[%s] clusterLabel doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, label)
		}
	}
}

func TestResourceLabel(t *testing.T) {
	testCases := []struct {
		name     string
		kind     string
		resName  string
		expected string
	}{
		{
			name:     "kind=pod and name=pod1 is specified",
			kind:     "pod",
			resName:  "pod1",
			expected: "<<TABLE BORDER=\"0\"><TR><TD><IMG SRC=\"/testdir/icons/pod-128.png\" /></TD></TR><TR><TD>pod1</TD></TR></TABLE>>",
		},
		{
			name:     "kind=svc and name=svc1 is specified",
			kind:     "svc",
			resName:  "svc1",
			expected: "<<TABLE BORDER=\"0\"><TR><TD><IMG SRC=\"/testdir/icons/svc-128.png\" /></TD></TR><TR><TD>svc1</TD></TR></TABLE>>",
		},
	}

	g := prepTestGraph(t)
	for _, tc := range testCases {
		label := g.resourceLabel(tc.kind, tc.resName)
		if tc.expected != label {
			t.Fatalf("[%s] resourceLabel doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, label)
		}
	}
}

func TestEscapeName(t *testing.T) {
	testCases := []struct {
		name     string
		resName  string
		expected string
	}{
		{
			name:     "Name without . and -, returns the same name",
			resName:  "my_namespace",
			expected: "my_namespace",
		},
		{
			name:     "Name with ., returns . replaced with _",
			resName:  "my.namespace",
			expected: "my_namespace",
		},
		{
			name:     "Name with -, returns . replaced with _",
			resName:  "my-namespace",
			expected: "my_namespace",
		},
		{
			name:     "Name with multiple - and ., returns all - and . replaced with _",
			resName:  "my-name.space_with.multiple.and-",
			expected: "my_name_space_with_multiple_and_",
		},
	}

	g := prepTestGraph(t)
	for _, tc := range testCases {
		name := g.escapeName(tc.resName)
		if tc.expected != name {
			t.Fatalf("[%s] escapeName doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, name)
		}
	}
}

func TestResourceName(t *testing.T) {
	testCases := []struct {
		name     string
		kind     string
		resName  string
		expected string
	}{
		{
			name:     "kind=pod and name=pod1 is specified",
			kind:     "pod",
			resName:  "pod1",
			expected: "pod_pod1",
		},
		{
			name:     "kind=svc and name=svc1 is specified",
			kind:     "svc",
			resName:  "svc1",
			expected: "svc_svc1",
		},
	}

	g := prepTestGraph(t)
	for _, tc := range testCases {
		name := g.resourceName(tc.kind, tc.resName)
		if tc.expected != name {
			t.Fatalf("[%s] resourceName doesn't return expected, expected:%v, returned:%v", tc.name, tc.expected, name)
		}
	}
}
