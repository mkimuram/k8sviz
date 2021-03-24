#! /bin/bash

KEEP_RESULT="${KEEP_RESULT:-false}"

BIN_NAME="k8sviz"
KIND_URL="https://kind.sigs.k8s.io/dl/v0.10.0/kind-linux-amd64"

E2E_DIR="$(dirname ${BASH_SOURCE})/"
TOP_DIR="${E2E_DIR}../../"
BIN_DIR="${TOP_DIR}bin/"
BIN_PATH="${BIN_DIR}${BIN_NAME}"
ICON_DIR="${TOP_DIR}icons/"
DATA_DIR="${TOP_DIR}test/data/"

if [ -n "${USE_EXISTING_DIR}" ];then
	TMP="${USE_EXISTING_DIR}"
else
	# Creat temp dir for test
	TMP="$(mktemp -d -p /tmp k8sviz-temp.XXXXXXXX)"
	if [ $? -ne 0 ];then
		echo "Failed to create temp dir for test"
		exit 1
	fi
fi

# Ensure that ${TMP} is not "", to avoid / to be deleted
if [ -z "${TMP}" ];then
	echo "Created temp dir name was empty"
	exit 1
fi

TEST_DIR="${TMP}/"
TEST_BIN_DIR="${TEST_DIR}bin/"
TEST_BIN_PATH="${TEST_BIN_DIR}${BIN_NAME}"
TEST_KIND_BIN="${TEST_DIR}/kind"
TEST_KUBECONFIG_PATH="${TEST_DIR}/kubeconfig"

# Create cluster name based on ${TMP} (take last 8 letters and make it lowercase and use it as suffix).
NAME_BASE=$(basename ${TMP})
CLUSTER_NAME="k8sviz-$(echo ${NAME_BASE: -8} | tr '[:upper:]' '[:lower:]')"

### Functions
function cleanup() {
	rm -rf "${TEST_DIR}"
}

function prepare() {
	# Check k8sviz binary exists in ${BIN_DIR}
	mkdir -p ${TEST_BIN_DIR}
	if [ $? -ne 0 ];then
		echo "Failed to make ${TEST_BIN_DIR} directory"
		cleanup
		exit 1
	fi

	cp ${BIN_PATH} ${TEST_BIN_PATH}
	if [ $? -ne 0 ];then
		echo "Failed to copy ${BIN_PATH} to ${TEST_BIN_PATH}"
		cleanup
		exit 1
	fi

	cp -r ${ICON_DIR} ${TEST_BIN_DIR}
	if [ $? -ne 0 ];then
		echo "Failed to copy ${ICON_DIR} to ${TEST_BIN_DIR}"
		cleanup
		exit 1
	fi
}

# create kind cluster
function create_cluster() {
	# Download kind
	curl -Lo ${TEST_KIND_BIN} https://kind.sigs.k8s.io/dl/v0.10.0/kind-linux-amd64
	chmod +x ${TEST_KIND_BIN}

	# Change path for kubeconfig 
	export KUBECONFIG=${TEST_KUBECONFIG_PATH}

	# Create cluster
	${TEST_KIND_BIN} create cluster --name ${CLUSTER_NAME}
}

function prepare_all() {
	if [ -n "${USE_EXISTING_DIR}" ];then
		echo "USE_EXISTING_DIR==${USE_EXISTING_DIR}, use it instead of creating new cluster"
	else
		prepare
		create_cluster
	fi
}

function delete_cluster() {
	# Change path for kubeconfig 
	export KUBECONFIG=${TEST_KUBECONFIG_PATH}

	# Delete cluster
	${TEST_KIND_BIN} delete cluster --name ${CLUSTER_NAME}
}

function cleanup_all() {
	if [ "${KEEP_RESULT}" == "true" ];then
		echo "KEEP_RESULT==true, keeping ${TEST_DIR} directory and ${CLUSTER_NAME} cluster"
	else
		delete_cluster
		cleanup
	fi
}

function cleanup_all_on_failure() {
	cleanup_all
	exit 1
}

# Run e2e tests
function run_test() {
	go test -v `go list ${E2E_DIR}...` -args -ginkgo.v \
		-kubeconfig ${TEST_KUBECONFIG_PATH} \
		-testdir ${TEST_DIR} \
		-testbin ${TEST_BIN_PATH} \
		-testdata ${DATA_DIR} \
	|| cleanup_all_on_failure
}

### Main
prepare_all
run_test
cleanup_all
