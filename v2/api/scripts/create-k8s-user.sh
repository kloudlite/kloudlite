#!/bin/bash
set -e

# Script to create a Kubernetes user with certificate-based authentication and RBAC

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
USER_NAME="${1:-kloudlite-api}"
NAMESPACE="${2:-kloudlite-system}"
CLUSTER_NAME="${3:-k3s-default}"
OUTPUT_DIR="${4:-./kubeconfig}"

echo -e "${GREEN}Creating Kubernetes user: ${USER_NAME}${NC}"

# Create output directory
mkdir -p ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}/certs

# Generate private key for the user
echo -e "${YELLOW}Generating private key...${NC}"
openssl genrsa -out ${OUTPUT_DIR}/certs/${USER_NAME}.key 2048

# Create a certificate signing request (CSR)
echo -e "${YELLOW}Creating certificate signing request...${NC}"
cat > ${OUTPUT_DIR}/certs/${USER_NAME}-csr.conf <<EOF
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
[ dn ]
CN = ${USER_NAME}
O = platform:admins
[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment,digitalSignature
extendedKeyUsage=clientAuth
EOF

openssl req -config ${OUTPUT_DIR}/certs/${USER_NAME}-csr.conf -new -key ${OUTPUT_DIR}/certs/${USER_NAME}.key -out ${OUTPUT_DIR}/certs/${USER_NAME}.csr

# Create CertificateSigningRequest in Kubernetes
echo -e "${YELLOW}Creating Kubernetes CertificateSigningRequest...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: ${USER_NAME}-csr
spec:
  request: $(cat ${OUTPUT_DIR}/certs/${USER_NAME}.csr | base64 | tr -d '\n')
  signerName: kubernetes.io/kube-apiserver-client
  usages:
  - client auth
EOF

# Approve the CSR
echo -e "${YELLOW}Approving certificate signing request...${NC}"
kubectl certificate approve ${USER_NAME}-csr

# Wait for certificate to be issued
echo -e "${YELLOW}Waiting for certificate...${NC}"
sleep 2

# Get the signed certificate
kubectl get csr ${USER_NAME}-csr -o jsonpath='{.status.certificate}' | base64 -d > ${OUTPUT_DIR}/certs/${USER_NAME}.crt

# Create namespace if it doesn't exist
echo -e "${YELLOW}Creating namespace ${NAMESPACE}...${NC}"
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Create RBAC roles for the user
echo -e "${YELLOW}Creating RBAC roles...${NC}"
cat <<EOF | kubectl apply -f -
---
# ClusterRole for managing platform CRDs
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${USER_NAME}-platform-admin
rules:
# User CRD permissions
- apiGroups: ["platform.kloudlite.io"]
  resources: ["users"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["platform.kloudlite.io"]
  resources: ["users/status"]
  verbs: ["get", "patch", "update"]
# Core resources needed by the API
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list", "create"]
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
# ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${USER_NAME}-platform-admin-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ${USER_NAME}-platform-admin
subjects:
- kind: User
  name: ${USER_NAME}
  apiGroup: rbac.authorization.k8s.io
EOF

# Get cluster CA certificate
echo -e "${YELLOW}Getting cluster CA certificate...${NC}"
kubectl config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d > ${OUTPUT_DIR}/certs/ca.crt

# Get cluster server URL
CLUSTER_SERVER=$(kubectl config view --raw -o jsonpath='{.clusters[0].cluster.server}')

# Create kubeconfig for the user
echo -e "${YELLOW}Creating kubeconfig file...${NC}"
cat > ${OUTPUT_DIR}/${USER_NAME}-kubeconfig.yaml <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $(cat ${OUTPUT_DIR}/certs/ca.crt | base64 | tr -d '\n')
    server: ${CLUSTER_SERVER}
  name: ${CLUSTER_NAME}
contexts:
- context:
    cluster: ${CLUSTER_NAME}
    user: ${USER_NAME}
    namespace: ${NAMESPACE}
  name: ${USER_NAME}@${CLUSTER_NAME}
current-context: ${USER_NAME}@${CLUSTER_NAME}
users:
- name: ${USER_NAME}
  user:
    client-certificate-data: $(cat ${OUTPUT_DIR}/certs/${USER_NAME}.crt | base64 | tr -d '\n')
    client-key-data: $(cat ${OUTPUT_DIR}/certs/${USER_NAME}.key | base64 | tr -d '\n')
EOF

# Clean up CSR from cluster
echo -e "${YELLOW}Cleaning up CSR...${NC}"
kubectl delete csr ${USER_NAME}-csr

echo -e "${GREEN}✅ User ${USER_NAME} created successfully!${NC}"
echo -e "${GREEN}Kubeconfig saved to: ${OUTPUT_DIR}/${USER_NAME}-kubeconfig.yaml${NC}"
echo ""
echo -e "${YELLOW}Test the connection with:${NC}"
echo "kubectl --kubeconfig=${OUTPUT_DIR}/${USER_NAME}-kubeconfig.yaml get users"
echo ""
echo -e "${YELLOW}Use this kubeconfig in your application:${NC}"
echo "export KUBECONFIG=${OUTPUT_DIR}/${USER_NAME}-kubeconfig.yaml"