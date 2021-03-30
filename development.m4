### merge dot files
digraph k8s_diagram {
define(`digraph',`subgraph')
include(out/gitlab-managed-apps.dot)dnl
include(out/jmeter.dot)dnl
}
