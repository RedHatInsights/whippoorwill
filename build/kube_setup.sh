#!/bin/bash
set -e

# Script you can use to set up a local minikube cluster for testing
# It is assumed you have already run 'minikube start' and your kubectl context is using the minikube cluster

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
    STRIMZI_VERSION=0.22.1
    STRIMZI_OPERATOR_NS=strimzi
    WATCH_NS="*"
    STRIMZI_TARFILE="strimzi-${STRIMZI_VERSION}.tar.gz"

    if [ $REINSTALL -ne 1 ]; then
        STRIMZI_DEPLOYMENT=$(kubectl get deployment strimzi-cluster-operator -n $STRIMZI_OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
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
    kubectl create namespace $STRIMZI_OPERATOR_NS || echo " ... ignoring that error"

    echo "*** Adding cluster-wide RoleBindings for Strimzi ..."
    # if we hit an error, assumption is the ClusterRoleBinding already exists
    kubectl create clusterrolebinding strimzi-cluster-operator-namespaced \
        --clusterrole=strimzi-cluster-operator-namespaced --serviceaccount ${STRIMZI_OPERATOR_NS}:strimzi-cluster-operator || echo " ... ignoring that error"
    kubectl create clusterrolebinding strimzi-cluster-operator-entity-operator-delegation \
        --clusterrole=strimzi-entity-operator --serviceaccount ${STRIMZI_OPERATOR_NS}:strimzi-cluster-operator || echo " ... ignoring that error"
    kubectl create clusterrolebinding strimzi-cluster-operator-topic-operator-delegation \
        --clusterrole=strimzi-topic-operator --serviceaccount ${STRIMZI_OPERATOR_NS}:strimzi-cluster-operator || echo " ... ignoring that error"

    if [ $REINSTALL -ne 1 ]; then
        echo "*** Installing Strimzi resources ..."
        kubectl create -f . -n $STRIMZI_OPERATOR_NS
    else
        echo "*** Replacing Strimzi resources ..."
        kubectl replace -f . -n $STRIMZI_OPERATOR_NS
    fi

    echo "*** Will wait for Strimzi operator to come up in background"
    kubectl rollout status deployment/strimzi-cluster-operator -n $STRIMZI_OPERATOR_NS | sed "s/^/[strimzi] /" &
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
    kubectl apply -f cert-manager.yaml

    echo "*** Will wait for cert manager to come up in background"
    kubectl rollout status deployment/cert-manager -n cert-manager | sed "s/^/[cert-manager] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}

function install_prometheus_operator {
    PROM_VERSION=0.45.0
    PROM_OPERATOR_NS=default
    PROM_TARFILE="prometheus-operator-${PROM_VERSION}.tar.gz"

    if [ $REINSTALL -ne 1 ]; then
        PROM_DEPLOYMENT=$(kubectl get deployment prometheus-operator -n $PROM_OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
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
    kubectl apply -f bundle.yaml

    echo "*** Will wait for Prometheus operator to come up in background"
    kubectl rollout status deployment/prometheus-operator -n $PROM_OPERATOR_NS | sed "s/^/[prometheus] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}


function install_cyndi_operator {
    OPERATOR_NS=cyndi-operator
    DEPLOYMENT=cyndi-operator-controller-manager

    if [ $REINSTALL -ne 1 ]; then
        OPERATOR_DEPLOYMENT=$(kubectl get deployment $DEPLOYMENT -n $OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
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
    kubectl apply -f cyndi-operator-manifest.yaml

    echo "*** Will wait for cyndi-operator to come up in background"
    kubectl rollout status deployment/$DEPLOYMENT -n $OPERATOR_NS | sed "s/^/[cyndi-operator] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}

function install_xjoin_operator {
    OPERATOR_NS=xjoin-operator-system
    DEPLOYMENT=xjoin-operator-controller-manager

    if [ $REINSTALL -ne 1 ]; then
        OPERATOR_DEPLOYMENT=$(kubectl get deployment $DEPLOYMENT -n $OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
        if [ ! -z "$OPERATOR_DEPLOYMENT" ]; then
            echo "*** xjoin-operator deployment found, skipping install ..."
            return 0
        fi
    fi

    echo "*** Installing xjoin-operator ..."
    cd "$DOWNLOAD_DIR"

    echo "*** Looking up latest release ..."
    LATEST_MANIFEST=$(curl -sL https://api.github.com/repos/RedHatInsights/xjoin-operator/releases/latest | jq -r '.assets[].browser_download_url')
    echo "*** Downloading $LATEST_MANIFEST ..."
    curl -LsS $LATEST_MANIFEST -o xjoin-operator-manifest.yaml

    echo "*** Applying xjoin-operator manifest ..."
    kubectl apply -f xjoin-operator-manifest.yaml

    echo "*** Will wait for xjoin-operator to come up in background"
    kubectl rollout status deployment/$DEPLOYMENT -n $OPERATOR_NS | sed "s/^/[xjoin-operator] /" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}

function install_elasticsearch_operator {
    OPERATOR_NS=elastic-system
    POD=elastic-operator-0

    if [ $REINSTALL -ne 1 ]; then
        OPERATOR_POD=$(kubectl get pod $POD -n $OPERATOR_NS --ignore-not-found -o jsonpath='{.metadata.name}')
        if [ ! -z "$OPERATOR_POD" ]; then
            echo "*** elastic-operator-0 pod found, skipping install ..."
            return 0
        fi
    fi

    echo "*** Applying elastic-operator manifest ..."
    kubectl apply -f https://download.elastic.co/downloads/eck/1.6.0/all-in-one.yaml

    echo "*** Will wait for elastic-operator to come up in background"
    kubectl wait pods/$POD --for=condition=Ready --timeout=150s -n "$OPERATOR_NS" &
    BG_PIDS+=($!)

    cd "$ROOT_DIR"
}

install_strimzi_operator
install_cert_manager
install_prometheus_operator
install_cyndi_operator
install_xjoin_operator
install_elasticsearch_operator

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
