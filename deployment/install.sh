#!/bin/bash
set -e
set -o pipefail

[ "$1" == "-x" ] && set -x && shift

SERVICE_ACCOUNT_NAME="kelefstis"
NAMESPACE="kube-system"

TMP=$(mktemp -d /tmp/k7s.XXXXXXXXXX)


KUBECFG="$TMP/config"

# if not present
# create ${SERVICE_ACCOUNT_NAME} in ${NAMESPACE} namespace
kubectl get sa  "${SERVICE_ACCOUNT_NAME}" --namespace "${NAMESPACE}" || \
kubectl create sa "${SERVICE_ACCOUNT_NAME}" --namespace "${NAMESPACE}" 

# Getting secret of service account ${SERVICE_ACCOUNT_NAME} on ${NAMESPACE}
SECRET_NAME=$(kubectl get sa "${SERVICE_ACCOUNT_NAME}" --namespace="${NAMESPACE}" -o jsonpath='{.secrets[0].name}' )


# Extracting ca.crt from secret...
kubectl get secret --namespace "${NAMESPACE}" "${SECRET_NAME}" -o jsonpath='{.data.ca\.crt}' | base64 --decode > "${TMP}/ca.crt"


# Getting user token from secret...
USER_TOKEN=$(kubectl get secret --namespace "${NAMESPACE}" "${SECRET_NAME}" -o jsonpath='{.data.token}' | base64 --decode)


# get current context
CONTEXT=$(kubectl config current-context)


CLUSTER_NAME=$(kubectl config get-contexts "$CONTEXT" | awk '{print $3}' | tail -n 1)
# "Cluster name: ${CLUSTER_NAME}"

ENDPOINT=$(kubectl config view \
-o jsonpath="{.clusters[?(@.name == \"${CLUSTER_NAME}\")].cluster.server}")
# "Endpoint: ${ENDPOINT}"

# Set up the config
# Preparing k8s-${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-conf"
# Setting a cluster entry in kubeconfig...
kubectl config set-cluster "${CLUSTER_NAME}" \
  --kubeconfig="${KUBECFG}" \
  --server="${ENDPOINT}" \
  --certificate-authority="${TMP}/ca.crt" \
  --embed-certs=true

# Setting token credentials entry in kubeconfig...
kubectl config set-credentials \
  "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
  --kubeconfig="${KUBECFG}" \
  --token="${USER_TOKEN}"

# Setting a context entry in kubeconfig...
kubectl config set-context \
  "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
  --kubeconfig="${KUBECFG}" \
  --cluster="${CLUSTER_NAME}" \
  --user="${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
  --namespace="${NAMESPACE}"

# setting the current-context in the kubeconfig file...
kubectl config use-context "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
  --kubeconfig="${KUBECFG}"

kubectl create secret -n kube-system generic kelefstis --from-file=$TMP/config || true

# create roles
kubectl apply -f role.yaml

# create the rulechecker crd and the rules
kubectl apply -f ../artifacts/examples/rulecheckers-crd.yaml || true
kubectl apply -f ../artifacts/examples/rules.all.yaml  || true

# create kelefstis logger
kubectl apply -f k7s.yaml

while ! kubectl logs  -n kube-system -l run=kelefstis ; do sleep 5; done
