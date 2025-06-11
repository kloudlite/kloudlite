#! /usr/bin/env bash

# this script is meant to be used in remote cluster's master, to get a k8s kubeconfig.yaml,
# that, can be de-commissioned in case it has been leaked.

# the generated kubeconfig, can be used to generate other kubeconfigs, and de-commissioning it would be as simple as:
# - deleting the service account
# - deleting the secret associated with that service account

output_file=$1
[ -z "$output_file" ] && echo "output_file must be defined as 1st argument to script, exiting ..." && exit 1

username="kloudlite-byok"
namespace="default"

echo "env-var KUBECTL is set to: $KUBECTL"

# if env var KUBECTL is defined, use it, else use kubectl executable from PATH
KUBECTL="${KUBECTL:-kubectl}"

echo "KUBECTL is set to: $KUBECTL"

$KUBECTL config view

cluster_name="byok"
cluster_addr=$(kubectl cluster-info | grep -i 'control plane' | awk '{print $7'})

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

cat >logs-reader-role.yml <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ${svc_account_name}-logs-reader
  namespace: ${namespace}
rules:
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list", "create"]
EOF

# cluster role binding to this user
cat >cluster-role-binding.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${svc_account_name}-cluster-rb
  namespace: ${namespace}
subjects:
  - kind: ServiceAccount
    name: ${svc_account_name}
    namespace: ${namespace}
roleRef:
  kind: Role
  name: ${svc_account_name}-logs-reader
  apiGroup: "rbac.authorization.k8s.io"
EOF

# while loop, as sometimes, when bootstrapping clusters, k3s takes sometime before responding
while true; do
	$KUBECTL apply -f .

	$KUBECTL get sa/${svc_account_name} -n ${namespace} >/dev/null 2>&1
	exit_status=$?
	if [ $exit_status -ne 0 ]; then
		continue
	fi

	$KUBECTL get secret/${svc_account_name}-token-secret -n ${namespace} >/dev/null 2>&1
	exit_status=$?
	if [ $exit_status -ne 0 ]; then
		continue
	fi
	break
done

popd || exit 1

# Get service account CA cert from secret
cert=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o jsonpath={.data."ca\.crt"})

# Get service account token from secret
user_token=""
while [ -z "$user_token" ]; do
	# sometimes it takes time for kubernetes controllers to reconcile service account secret
	user_token=$($KUBECTL get secret "${svc_account_secret_name}" -n "${namespace}" -o jsonpath={.data."token"} | base64 -d)
done

### now generating a new kubeconfig from this generated service account token
cat > "${output_file}" <<EOF
apiVersion: v1
kind: Config
clusters:
  - name: ${cluster_name}
    cluster:
      certificate-authority-data: ${cert}
      server: ${cluster_addr}
contexts:
  - name: ${new_context_name}
    context:
      cluster: ${cluster_name}
      namespace: ${namespace}
      user: ${svc_account_name}
users:
  - name: ${svc_account_name}
    user:
      token: ${user_token}
current-context: ${new_context_name}
EOF

echo "certificate is valid now, generated ${output_file} can now be used to communicate with the cluster"
echo "DONE 🚀"
