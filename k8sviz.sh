#! /bin/bash

#### Variables ####
NAME=$(basename $0 | tr - ' ')
NAMESPACE="default"
OUTFILE="k8sviz.out"
TYPE="dot"
KUBECONFIG=~/kubeconfig
CONTAINER_IMG=mkimuram/k8sviz:0.2

#### Functions ####
function help () {
  cat << EOF
Generate Kubernetes architecture diagrams from the actual state in a namespace
Usage:
  $NAME [options]
Options:
  -h, --help                 Displays the help text
  -n, --namespace            The namespace to visualize. Default is ${NAMESPACE}
  -o, --outfile              The filename to output. Default is ${OUTFILE}
  -t, --type                 The type of output. Default is ${TYPE}
  -k, --kubeconfig           Path to kubeconfig file. Default is ${KUBECONFIG}
  -i, --image                Image name of the container. Default is ${CONTAINER_IMG}
EOF
}

#### Main ####
# Parse Options
OPTS=$(getopt --options hn:o:t:k:i: --longoptions help,namespace:,outfile:,type:,kubeconfig:,image: --name "$NAME" -- "$@")
[[ $? != 0 ]] && echo "Failed parsing options" >&2 && exit 1
eval set -- "$OPTS"

while true;do
  case "$1" in
    -h | --help)
      help
      exit 0
      ;;
    -n | --namespace)
      NAMESPACE="${2:-$NAMESPACE}"
      shift 2
      ;;
    -o | --outfile)
      OUTFILE="${2:-$OUTFILE}"
      shift 2
      ;;
    -t | --type)
      TYPE="${2:-$TYPE}"
      shift 2
      ;;
    -k | --kubeconfig)
      KUBECONFIG="${2:-$KUBECONFIG}"
      shift 2
      ;;
    -i | --image)
      CONTAINER_IMG="${2:-$CONTAINER_IMG}"
      shift 2
      ;;
    --)
      shift
      break
      ;;
  esac
done

# Split OUTFILE to the directory and the filename to be used with container
DIR=$(dirname ${OUTFILE})
ABSDIR=$(cd ${DIR}; pwd -P)
FILENAME=$(basename ${OUTFILE})

# Check if KUBECONFIG file exists
if [ ! -f "${KUBECONFIG}" ];then
  echo "KUBECONFIG file wasn't found in ${KUBECONFIG}." >&2
  echo "You may need to specify the path with --kubeconfig option." >&2
  exit 1
fi

docker run --network host -v ${ABSDIR}:/work -v ${KUBECONFIG}:/config:ro \
  -e KUBECONFIG=/config -it ${CONTAINER_IMG} /k8sviz -kubeconfig /config \
  -n ${NAMESPACE} -t ${TYPE} -o /work/${FILENAME}
