#! /usr/bin/env bash

# this script is meant to be used in remote cluster's master, to get a k8s kubeconfig.yaml,
# that, can be de-commissioned in case it has been leaked.

# the generated kubeconfig, can be used to generate other kubeconfigs, and de-commissioning it would be as simple as:
# - deleting the service account
# - deleting the secret associated with that service account

output_file=$1
[ -z "$output_file" ] && echo "output_file must be defined as 1st argument to script, exiting ..." && exit 1

username="kloudlite-admin"
namespace="default"

echo "env-var KUBECTL is set to: $KUBECTL"

# if env var KUBECTL is defined, use it, else use kubectl executable from PATH
KUBECTL="${KUBECTL:-kubectl}"

echo "KUBECTL is set to: $KUBECTL"

curr_context_name=$($KUBECTL config view -o jsonpath='{.current-context}')
cluster_name=$($KUBECTL config view -o jsonpath="{.contexts[?(@.name=='${curr_context_name}')].context.cluster}")

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

### now generating a new kubeconfig from this generated service account token
function generate_kubeconfig() {
	token=$1
	output=$2

	$KUBECTL config set-credentials "${username}" --token="${token}"
	$KUBECTL config set-context ${new_context_name} --cluster="${cluster_name}" --user="${username}" --namespace="${namespace}"
	$KUBECTL config use-context ${new_context_name}

	echo "saving generated kubeconfig to kubeconfig.yml"
	$KUBECTL config view --raw --minify=true >"${output}"
	$KUBECTL config use-context "${curr_context_name}"

	echo "cleaning up existing kubeconfig for sanity reasons ..."
	$KUBECTL config delete-context ${new_context_name}
	$KUBECTL config delete-user "${username}"
}

# Get service account token from secret
# user_token=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o json | jq -r '.data["token"]' | base64 -d)
user_token=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o jsonpath={.data."token"} | base64 -d)
generate_kubeconfig "${user_token}" "${output_file}"

cert=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o jsonpath={.data."ca\.crt"} | base64 -d)

validity_starts_at=$(echo "${cert}" | openssl x509 -noout -startdate | awk -F= '{print $2}')
echo "${output_file} will be valid after ${validity_starts_at}"

validity_start_timestamp=$(date -d "${validity_starts_at}" +%s)
curr_timestamp="$(date +%s)"

echo "validity_start_timestamp: ${validity_start_timestamp}"
echo "curr_timestamp: ${curr_timestamp}"
diff=$((validity_start_timestamp - curr_timestamp))

echo "certificate will be valid in ${diff} seconds ..."
while [ $((diff)) -ge 0 ]; do
	echo "certificate will be valid in ${diff} seconds ..."
	sleep 1
	curr_timestamp="$(date +%s)"
	diff=$((validity_start_timestamp - curr_timestamp))
done

echo "certificate is valid now, generated ${output_file} can now be used to communicate with the cluster"
echo "DONE 🚀"
