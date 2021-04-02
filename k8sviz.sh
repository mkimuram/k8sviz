#! /bin/bash

#### Variables ####
NAMESPACE="default"
OUTDIR="out"
TYPE="dot"
GLOBAL_FILENAME=""
KUBECONFIG=~/kubeconfig
#CONTAINER_IMG=mkimuram/k8sviz:0.3
CONTAINER_IMG=docker.io/library/k8sviz:devel
SHFLAGS_DIR="$(dirname ${BASH_SOURCE})/lib/"
SHFLAGS_PATH="${SHFLAGS_DIR}shflags"
SHFLAGS_URL="https://raw.githubusercontent.com/kward/shflags/master/shflags"

if [ ! -f ${SHFLAGS_PATH} ];then
	echo "${SHFLAGS_PATH} not found. Downloading." >&2

	mkdir -p ${SHFLAGS_DIR}
	if [ $? -ne 0 ];then
		cat << EOF >&2
Failed to create ${SHFLAGS_DIR} directory.
Move this script to the directory where you have write permission.
EOF
		exit 1
	fi

	curl -L -f -o ${SHFLAGS_PATH} ${SHFLAGS_URL}
	if [ $? -ne 0 ];then
		cat << EOF >&2
Failed to download shflags.
You can manually download it from ${SHFLAGS_URL}
and copy it to ${SHFLAGS_DIR} to fix it.
EOF
		exit 1
	fi
fi

. ${SHFLAGS_PATH}

DEFINE_string 'namespace' "${NAMESPACE}" 'The namespace to visualize.' 'n'
DEFINE_string 'outdir' "${OUTDIR}" 'The directory to output.' 'o'
DEFINE_string 'type' "${TYPE}" 'The type of output.' 't'
DEFINE_string 'kubeconfig' "${KUBECONFIG}" 'Path to kubeconfig file.' 'k'
DEFINE_string 'image' "${CONTAINER_IMG}" 'Image name of the container.' 'i'
DEFINE_string 'filename' "${GLOBAL_FILENAME}" 'global filename with concatenated diagram.' 'f'

# Parse Options
FLAGS "$@" || exit $?
eval set -- "${FLAGS_ARGV}"

#### Main ####
# Split OUTDIR to the directory and the filename to be used with container
DIR=$(dirname ${FLAGS_outdir})
ABSDIR=$(cd ${DIR}; pwd -P)
FILENAME=$(basename ${FLAGS_outdir})

# Make KUBECONFIG to absolute path
KUBEDIR=$(dirname ${FLAGS_kubeconfig})
ABSKUBEDIR=$(cd ${KUBEDIR}; pwd -P)
KUBEFILE=$(basename ${FLAGS_kubeconfig})
KUBECONFIG="${ABSKUBEDIR}/${KUBEFILE}"

# Check if KUBECONFIG file exists
if [ ! -f "${KUBECONFIG}" ];then
  echo "KUBECONFIG file wasn't found in ${KUBECONFIG}." >&2
  echo "You need to specify the right path with --kubeconfig option." >&2
  exit 1
fi

if [ -z $GLOBAL_FILENAME ]; then
  docker run --network host                                            \
    --user $(id -u):$(id -g)                                           \
    -v ${ABSDIR}:/work                                                 \
    -v ${KUBECONFIG}:/config:ro                                        \
    -e GOOGLE_APPLICATION_CREDENTIALS=/work/service-account-key.json   \
    -it --rm ${FLAGS_image}                                            \
    /k8sviz -kubeconfig /config                                        \
    -t ${FLAGS_type} -n ${FLAGS_namespace} -o /work/${FILENAME}

else
  docker run --network host                                            \
    --user $(id -u):$(id -g)                                           \
    -v ${ABSDIR}:/work                                                 \
    -v ${KUBECONFIG}:/config:ro                                        \
    -e GOOGLE_APPLICATION_CREDENTIALS=/work/service-account-key.json   \
    -it --rm ${FLAGS_image}                                            \
    /k8sviz -kubeconfig /config                                        \
    -n ${FLAGS_namespace} -t 'dot' -o /work/${FILENAME}
	m4 merge.m4 > merged.gv
	sed -i -e "s/\/icons/icons/" merged.gv
	if [ "dot" != "${FLAGS_type}" ]; then
	  dot -n -T${FLAGS_type} merged.gv -o $GLOBAL_FILENAME.${FLAGS_type}
	  rm merged.gv
	else
	  mv merged.gv $GLOBAL_FILENAME.${FLAGS_type}
	fi
fi
