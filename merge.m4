### merge dot files
digraph k8s_diagram {
define(`digraph',`subgraph')
include(out/ns1.dot)dnl
include(out/ns2.dot)dnl
}
