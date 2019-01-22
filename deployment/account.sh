#!/bin/bash
set -e
set -o pipefail

[ $1 == "-x" ] && set -x && shift

[ $# -eq 2 ] || (echo "Usage $0 serviceaccount namespace" && exit 1)

SERVICE_ACCOUNT_NAME=$1
NAMESPACE=$2

TMP=$(mktemp -d /tmp/k7s.XXXXXXXXXX)


KUBECFG="$TMP/config"

# if not present
# create ${SERVICE_ACCOUNT_NAME} in ${NAMESPACE} namespace
kubectl get sa  "${SERVICE_ACCOUNT_NAME}" --namespace "${NAMESPACE}" || \
kubectl create sa "${SERVICE_ACCOUNT_NAME}" --namespace "${NAMESPACE}" 

# Getting secret of service account ${SERVICE_ACCOUNT_NAME} on ${NAMESPACE}
SECRET_NAME=$(kubectl get sa "${SERVICE_ACCOUNT_NAME}" --namespace="${NAMESPACE}" -o jsonpath='{.secrets[0].name}' )


# Extracting ca.crt from secret...
kubectl get secret --namespace "${NAMESPACE}" "${SECRET_NAME}" -o json | jq \
    -r '.data["ca.crt"]' | base64 --decode > "${TMP}/ca.crt"


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

# -n "Setting the current-context in the kubeconfig file..."
kubectl config use-context "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
  --kubeconfig="${KUBECFG}"

# -e "\\nAll done! Test with:"
# "KUBECONFIG=${KUBECFG} kubectl get pods"
# "you should not have any permissions by default - you have just created the authentication part"
# "You will need to create RBAC permissions"
KUBECONFIG=${KUBECFG} kubectl get pods

echo "create the secret"
echo kubectl create secret -n kube-system generic kelefstis --from-file=/tmp/k7s.C6UbYfsFJ9/config
