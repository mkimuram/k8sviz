### merge dot files
digraph k8s_diagram {
define(`digraph',`subgraph')
include(out/postgresql-staging.dot)dnl
include(out/keycloak-staging.dot)dnl
include(out/mongodb-staging.dot)dnl
include(out/feamzy-web-staging.dot)dnl
include(out/feamzy-api-staging.dot)dnl
}
