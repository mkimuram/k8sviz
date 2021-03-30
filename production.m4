### merge dot files
digraph k8s_diagram {
define(`digraph',`subgraph')
include(out/postgresql-prod.dot)dnl
include(out/keycloak.dot)dnl
include(out/mongodb-production.dot)dnl
include(out/feamzy-web-production.dot)dnl
include(out/feamzy-api-production.dot)dnl
include(out/haproxy.dot)dnl
}
