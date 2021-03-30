### merge dot files
digraph k8s_diagram {
define(`digraph',`subgraph')
include(out/elasticsearch.dot)dnl
include(out/grafana-ds-api-plugin-25435752-production.dot)dnl
include(out/prometheus.dot)dnl
}
