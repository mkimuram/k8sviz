### merge dot files
digraph k8s_diagram {
define(`digraph',`subgraph')
include(out/mongodb.dot)dnl
include(out/postgresql.dot)dnl
}
