# k8sviz

k8sviz is a tool to generate Kubernetes architecture diagrams from the actual state in a namespace.
Currently, this only generates a diagram similar to https://github.com/kubernetes/community/tree/master/icons#usage-example by using [graphviz](https://www.graphviz.org/).
For examples of the generated diagrams, see [Examples](#examples) below.

## Implementations
There are two implementations, bash script version and go version. Bash script version was intended to be just a prototype. So, it will be migrated to go version.

## Prerequisites
### Bash script version
`k8sviz.sh` depends on below commands:
- awk
- bash
- cat
- dot (graphviz)
- getopt
- grep
- kubectl
- sed
- seq
- tr

### Go version
`k8sviz` only depends dot (graphviz) command.

## Installation
### Bash script version
Just git clone this repository or copy `k8sviz.sh` file and `icons` directory with keeping directory structure.

### Go version
```
$ git clone https://github.com/mkimuram/k8sviz.git
$ cd k8sviz
$ export GO111MODULE=on
$ go build -o k8sviz .
```

k8sviz binary can be moved to another directory, but `icons` directory needs to be in the same directory to the binary.

## Usage
### Bash script version
```
$ ./k8sviz.sh --help
Generate Kubernetes architecture diagrams from the actual state in a namespace
Usage:
  k8sviz.sh [options]
Options:
  -h, --help                 Displays the help text
  -n, --namespace            The namespace to visualize. Default is default
  -o, --outfile              The filename to output. Default is k8sviz.out
  -t, --type                 The type of output. Default is dot
```

### Go version
```
$ ./k8sviz -h
Usage of ./k8sviz:
  -kubeconfig string
        absolute path to the kubeconfig file (default "/root/.kube/config")
  -n string
        namespace to visualize (shorthand) (default "namespace")
  -namespace string
        namespace to visualize (default "namespace")
  -o string
        output filename (shorthand) (default "k8sviz.out")
  -outfile string
        output filename (default "k8sviz.out")
  -t string
        type of output (shorthand) (default "dot")
  -type string
        type of output (default "dot")
```

## Examples
Examples are only shown for bash script version, but go version should work in the same way.
Report bugs or critical differences, if you find any.

### Examples for tutorial deployments in default namespace
- Generate dot file for namespace `default`
```
$ ./k8sviz.sh -n default -o default.dot
```
- Generate png file for namespace `default`
```
$ ./k8sviz.sh -n default -t png -o default.png
```
- Output for [an example wordpress deployment](https://kubernetes.io/docs/tutorials/stateful-application/mysql-wordpress-persistent-volume/) will be like below:
   - [default.dot](./examples/wordpress/default.dot)
   - [default.png](./examples/wordpress/default.png):

<a href="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/wordpress/default.png"><img src="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/wordpress/default.png" width="40%" height="40%"/></a>
- Output for [an example cassandra deployment with statefulset](https://kubernetes.io/docs/tutorials/stateful-application/cassandra/) will be like below:
   - [default.dot](./examples/cassandra/default.dot)
   - [default.png](./examples/cassandra/default.png):

<a href="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/cassandra/default.png"><img src="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/cassandra/default.png" width="50%" height="50%"/></a>

### Examples for more complex deployment ([kubeflow](https://www.kubeflow.org/docs/started/k8s/kfctl-k8s-istio/) case)
- Generate dot file for namespace `kubeflow` and `istio-system`
```
$ ./k8sviz.sh -n kubeflow -o examples/kubeflow/kubeflow.dot
$ ./k8sviz.sh -n istio-system -o examples/kubeflow/istio-system.dot
```
- Generate png file for namespace `kubeflow` and `istio-system`
```
$ ./k8sviz.sh -n kubeflow -t png -o examples/kubeflow/kubeflow.png
$ ./k8sviz.sh -n istio-system -t png -o examples/kubeflow/istio-system.png
```
- Output:
   - [kubeflow.dot](./examples/kubeflow/kubeflow.dot)
   - [istio-system.dot](./examples/kubeflow/istio-system.dot)
   - [kubeflow.png](./examples/kubeflow/kubeflow.png)

   <a href="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/kubeflow/kubeflow.png"><img src="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/kubeflow/kubeflow.png" width="90%" height="90%"/></a>

   - [istio-system.png](./examples/kubeflow/istio-system.png)

   <a href="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/kubeflow/istio-system.png"><img src="https://raw.githubusercontent.com/mkimuram/k8sviz/master/examples/kubeflow/istio-system.png" width="90%" height="90%"/></a>

## License
This project is licensed under the Apache License - see the [LICENSE file](./LICENSE) for details
