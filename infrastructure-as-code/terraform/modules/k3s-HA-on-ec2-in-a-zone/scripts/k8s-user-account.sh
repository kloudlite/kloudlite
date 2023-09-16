#! /usr/bin/env bash

# this script is meant to be used in remote cluster's master, to get a k8s kubeconfig.yaml,
# that, can be de-commissioned in case it has been leaked.

# the generated kubeconfig, can be used to generate other kubeconfigs, and de-commissioning it would be as simple as:
# - deleting the service account
# - deleting the secret associated with that service account

username="kloudlite-admin"
namespace="default"

echo "env-var KUBECTL is set to: $KUBECTL"

# if env var KUBECTL is defined, use it, else use kubectl executable from PATH
KUBECTL="${KUBECTL:-kubectl}"

echo "KUBECTL is set to: $KUBECTL"

curr_context_name=$($KUBECTL config view -o jsonpath='{.current-context}')
# cluster_name=$($KUBECTL config view -o json | jq -r ".contexts[] | select(.name == \"$curr_context_name\")| .context.cluster")
cluster_name=$($KUBECTL config view -o jsonpath="{.contexts[?(@.name=='${curr_context_name}')].context.cluster}")
# cluster_url=$($KUBECTL config view -o json | jq -r ".clusters[] | select(.name == \"$curr_context_name\") | .cluster.server")
#cluster_url=$($KUBECTL config view -o jsonpath="{.clusters[?(@.name=='${cluster_name}')].cluster.server}")

new_context_name="${username}-ctx"

manifestsDir="$PWD/${username}/manifests"
[ -d "$manifestsDir" ] || mkdir -p "$manifestsDir"

pushd "$manifestsDir" || (echo "pushd failed, exiting ..." && exit 1)

svc_account_name="${username}"

# creating a new service account
cat >svc-account.yaml <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${svc_account_name}
  namespace: ${namespace}
EOF

$KUBECTL apply -f svc-account.yaml

svc_account_secret_name="${svc_account_name}-token-secret"

# creating a new service account secret
cat >svc-account-secret.yml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ${svc_account_secret_name}
  namespace: ${namespace}
  annotations:
    kubernetes.io/service-account.name: ${svc_account_name}
type: kubernetes.io/service-account-token
EOF

$KUBECTL apply -f svc-account-secret.yml

# cluster role binding to this user
cat >cluster-role-binding.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${svc_account_name}-cluster-rb
subjects:
  - kind: ServiceAccount
    name: ${svc_account_name}
    namespace: ${namespace}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: "rbac.authorization.k8s.io"
EOF

$KUBECTL apply -f cluster-role-binding.yaml
popd || exit 1

# Get service account token from secret
# user_token=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o json | jq -r '.data["token"]' | base64 -d)
user_token=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o jsonpath={.data."token"} | base64 -d)

### now generating a new kubeconfig from this generated service account token
#$KUBECTL config set-cluster "${cluster_name}" --embed-certs=true --server="${cluster_url}" --certificate-authority=/tmp/ca.crt
$KUBECTL config set-credentials "${username}" --token="${user_token}"
$KUBECTL config set-context ${new_context_name} --cluster="${cluster_name}" --user="${username}" --namespace="${namespace}"
$KUBECTL config use-context ${new_context_name}

echo "saving generated kubeconfig to kubeconfig.yml"
$KUBECTL config view --raw --minify=true >kubeconfig.yml
$KUBECTL config use-context "${curr_context_name}"

echo "cleaning up existing kubeconfig for sanity reasons ..."
$KUBECTL config delete-context ${new_context_name} >/dev/null
$KUBECTL config delete-user "${username}"

echo "DONE 🚀"
