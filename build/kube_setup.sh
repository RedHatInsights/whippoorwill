#!/bin/bash
# Configures a local minikube cluster for testing.
set -e

# Set to `kubectl` if kubectl is installed system-wide, its context is set to your minikube cluster,
# and you wish to use it instead of the kubectl bundled with minikube.
KUBECTL_CMD='minikube kubectl --'

REINSTALL=0

for arg in "$@"
do
    case $arg in
        -r|--reinstall)
        REINSTALL=1
        shift
        ;;
    esac
done

PLATFORM=`uname -a | cut -f1 -d' '`

# jq is required for cyndi operator install, check if jq is installed
echo "*** Checking for 'jq' ..."
if ! command -v jq; then
    echo "*** 'jq' not found in path. Please install jq with:"
    [[ $PLATFORM == "Darwin" ]] && echo "  brew install jq" \
        || echo "  sudo dnf install jq"
    exit 1
fi

# GO is required for yq, check if go is installed
echo "*** Checking for 'go' ..."
if ! command -v go; then
    echo "*** 'go' not found in path. Please install go with:"
    [[ $PLATFORM == "Darwin" ]] && echo "  'brew install golang' or the instructions at https://golang.org/doc/install" \
        || echo "  sudo dnf install golang"
    exit 1
fi


GO_BIN_PATH="$(go env GOPATH)/bin"

export PATH="$PATH:$GO_BIN_PATH"

echo "*** Checking for 'yq' ..."
if ! command -v yq; then
    echo "*** 'yq' executable not found in '$GO_BIN_PATH', installing it with:"
    echo "         GO111MODULE=on go get github.com/mikefarah/yq/v4"
    (cd & GO111MODULE=on go get github.com/mikefarah/yq/v4)
fi

declare -a array BG_PIDS=()

ROOT_DIR=$(pwd)
DOWNLOAD_DIR="build/operator_bundles"
mkdir -p "$DOWNLOAD_DIR"


function install_strimzi_operator {
    STRIMZI_VERSION=0.21.1
    STRIMZI_OPERATOR_NS=strimzi
    WATCH_NS="*"
    STRIMZI_TARFILE="strimzi-${STRIMZI_VERSION}.tar.gz"

    if [ $REINSTALL -ne 1 ]; then
        STRIMZI_DEPLOYMENT=$(${KUBECTL_CMD} get deployment strimzi-cluster-operator -n $STRIMZI_OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
        if [ ! -z "$STRIMZI_DEPLOYMENT" ]; then
            echo "*** Strimzi operator deployment found, skipping install ..."
            return 0
        fi
    fi

    echo "*** Installing strimzi operator ..."
    cd "$DOWNLOAD_DIR"

    if ! test -f ${STRIMZI_TARFILE}; then
        echo "*** Downloading ${STRIMZI_TARFILE} ..."
        curl -LsSO https://github.com/strimzi/strimzi-kafka-operator/releases/download/${STRIMZI_VERSION}/${STRIMZI_TARFILE}
    fi

    echo "*** Unpacking .tar.gz ..."
    tar xzf ${STRIMZI_TARFILE}

    echo "Setting namespaces (STRIMZI_OPERATOR_NS: $STRIMZI_OPERATOR_NS, WATCH_NS: $WATCH_NS) in strimzi configs ..."
    cd strimzi-${STRIMZI_VERSION}/install/cluster-operator
    # Set namespace that operator runs in
    [[ $PLATFORM == "Darwin" ]] && sed -i '' "s/namespace: .*/namespace: ${STRIMZI_OPERATOR_NS}/" *RoleBinding*.yaml \
        || sed -i "s/namespace: .*/namespace: ${STRIMZI_OPERATOR_NS}/" *RoleBinding*.yaml

    # Set namespaces that operator watches
    yq eval -i "del(.spec.template.spec.containers[0].env.[] | select(.name == \"STRIMZI_NAMESPACE\").valueFrom)" 060-Deployment-strimzi-cluster-operator.yaml
    yq eval -i "(.spec.template.spec.containers[0].env.[] | select(.name == \"STRIMZI_NAMESPACE\")).value = \"$WATCH_NS\"" 060-Deployment-strimzi-cluster-operator.yaml

    echo "*** Creating ns ${STRIMZI_OPERATOR_NS}..."
    # if we hit an error, assumption is the Namespace already exists
    ${KUBECTL_CMD} create namespace $STRIMZI_OPERATOR_NS || echo " ... ignoring that error"

    echo "*** Adding cluster-wide RoleBindings for Strimzi ..."
    # if we hit an error, assumption is the ClusterRoleBinding already exists
    ${KUBECTL_CMD} create clusterrolebinding strimzi-cluster-operator-namespaced \
        --clusterrole=strimzi-cluster-operator-namespaced --serviceaccount ${STRIMZI_OPERATOR_NS}:strimzi-cluster-operator || echo " ... ignoring that error"
    ${KUBECTL_CMD} create clusterrolebinding strimzi-cluster-operator-entity-operator-delegation \
        --clusterrole=strimzi-entity-operator --serviceaccount ${STRIMZI_OPERATOR_NS}:strimzi-cluster-operator || echo " ... ignoring that error"
    ${KUBECTL_CMD} create clusterrolebinding strimzi-cluster-operator-topic-operator-delegation \
        --clusterrole=strimzi-topic-operator --serviceaccount ${STRIMZI_OPERATOR_NS}:strimzi-cluster-operator || echo " ... ignoring that error"

    echo "*** Installing Strimzi resources ..."
    ${KUBECTL_CMD} apply -f . -n $STRIMZI_OPERATOR_NS

    echo "*** Will wait for Strimzi operator to come up in background"
    ${KUBECTL_CMD} rollout status deployment/strimzi-cluster-operator -n $STRIMZI_OPERATOR_NS | sed "s/^/[strimzi] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}

function install_cert_manager {
    CERT_MANAGER_VERSION=v1.2.0

    echo "*** Installing cert manager ..."
    cd "$DOWNLOAD_DIR"

    echo "*** Downloading ${CERT_MANAGER_YAML} ..."
    curl -LsSO https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml

    echo "*** Installing Cert Manager resources ..."
    ${KUBECTL_CMD} apply -f cert-manager.yaml

    echo "*** Will wait for cert manager to come up in background"
    ${KUBECTL_CMD} rollout status deployment/cert-manager -n cert-manager | sed "s/^/[cert-manager] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}

function install_prometheus_operator {
    PROM_VERSION=0.45.0
    PROM_OPERATOR_NS=default
    PROM_TARFILE="prometheus-operator-${PROM_VERSION}.tar.gz"

    if [ $REINSTALL -ne 1 ]; then
        PROM_DEPLOYMENT=$(${KUBECTL_CMD} get deployment prometheus-operator -n $PROM_OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
        if [ ! -z "$PROM_DEPLOYMENT" ]; then
            echo "*** Prometheus operator deployment found, skipping install ..."
            return 0
        fi
    fi

    echo "*** Installing prometheus operator ..."
    cd "$DOWNLOAD_DIR"

    if ! test -f ${PROM_TARFILE}; then
        echo "*** Downloading ${PROM_TARFILE} ..."
        curl -LsS -o ${PROM_TARFILE} https://github.com/prometheus-operator/prometheus-operator/archive/v${PROM_VERSION}.tar.gz
    fi

    echo "*** Unpacking .tar.gz ..."
    tar xzf ${PROM_TARFILE}

    echo "*** Applying prometheus operator manifest ..."
    cd prometheus-operator-${PROM_VERSION}
    ${KUBECTL_CMD} apply -f bundle.yaml

    echo "*** Will wait for Prometheus operator to come up in background"
    ${KUBECTL_CMD} rollout status deployment/prometheus-operator -n $PROM_OPERATOR_NS | sed "s/^/[prometheus] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}


function install_cyndi_operator {
    OPERATOR_NS=cyndi-operator
    DEPLOYMENT=cyndi-operator-controller-manager

    if [ $REINSTALL -ne 1 ]; then
        OPERATOR_DEPLOYMENT=$(${KUBECTL_CMD} get deployment $DEPLOYMENT -n $OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
        if [ ! -z "$OPERATOR_DEPLOYMENT" ]; then
            echo "*** cyndi-operator deployment found, skipping install ..."
            return 0
        fi
    fi

    echo "*** Installing cyndi-operator ..."
    cd "$DOWNLOAD_DIR"

    echo "*** Looking up latest release ..."
    LATEST_MANIFEST=$(curl -sL https://api.github.com/repos/RedHatInsights/cyndi-operator/releases/latest | jq -r '.assets[].browser_download_url')
    echo "*** Downloading $LATEST_MANIFEST ..."
    curl -LsS $LATEST_MANIFEST -o cyndi-operator-manifest.yaml

    echo "*** Applying cyndi-operator manifest ..."
    ${KUBECTL_CMD} apply -f cyndi-operator-manifest.yaml

    echo "*** Will wait for cyndi-operator to come up in background"
    ${KUBECTL_CMD} rollout status deployment/$DEPLOYMENT -n $OPERATOR_NS | sed "s/^/[cyndi-operator] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}


install_strimzi_operator
install_cert_manager
install_prometheus_operator
install_cyndi_operator

FAILURES=0
if [ ${#BG_PIDS[@]} -gt 0 ]; then
    echo "*** Waiting on background jobs to finish ..."
    for pid in ${BG_PIDS[*]}; do
        wait $pid || FAILURES+=1
    done
fi

if [ $FAILURES -gt 0 ]; then
    echo "*** ERROR: background job(s) failed"
    exit 1
else
    echo "*** Done!"
fi
