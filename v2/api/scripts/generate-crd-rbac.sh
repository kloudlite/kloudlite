#!/bin/bash
set -e

# Script to generate RBAC rules for a new CRD
# Usage: ./generate-crd-rbac.sh <crd-name> <api-group> [output-file]

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -lt 2 ]; then
    echo -e "${RED}Usage: $0 <crd-name> <api-group> [output-file]${NC}"
    echo -e "${YELLOW}Example: $0 teams platform.kloudlite.io rbac/teams-rbac.yaml${NC}"
    exit 1
fi

CRD_NAME=$1
API_GROUP=$2
OUTPUT_FILE=${3:-"rbac/${CRD_NAME}-rbac.yaml"}

# Ensure plural form
CRD_PLURAL="${CRD_NAME}"
if [[ ! "${CRD_NAME}" =~ s$ ]]; then
    CRD_PLURAL="${CRD_NAME}s"
fi

echo -e "${GREEN}Generating RBAC for CRD: ${CRD_NAME} (${API_GROUP})${NC}"

# Create output directory if needed
mkdir -p $(dirname ${OUTPUT_FILE})

# Generate RBAC YAML
cat > ${OUTPUT_FILE} <<EOF
---
# Auto-generated RBAC for ${CRD_NAME} CRD
# Generated on: $(date)

# ClusterRole for full admin access to ${CRD_NAME} resources
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: platform-admin-${CRD_PLURAL}
  labels:
    app.kubernetes.io/name: kloudlite
    app.kubernetes.io/component: rbac
    rbac.kloudlite.io/aggregate-to-platform-admin: "true"
    rbac.kloudlite.io/crd: ${CRD_NAME}
rules:
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}"]
  verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}/status"]
  verbs: ["get", "patch", "update"]
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}/finalizers"]
  verbs: ["update"]

---
# ClusterRole for read-only access to ${CRD_NAME} resources
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: platform-viewer-${CRD_PLURAL}
  labels:
    app.kubernetes.io/name: kloudlite
    app.kubernetes.io/component: rbac
    rbac.kloudlite.io/aggregate-to-platform-viewer: "true"
    rbac.kloudlite.io/crd: ${CRD_NAME}
rules:
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}/status"]
  verbs: ["get"]

---
# ClusterRole for editor access (no delete) to ${CRD_NAME} resources
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: platform-editor-${CRD_PLURAL}
  labels:
    app.kubernetes.io/name: kloudlite
    app.kubernetes.io/component: rbac
    rbac.kloudlite.io/aggregate-to-platform-editor: "true"
    rbac.kloudlite.io/crd: ${CRD_NAME}
rules:
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}"]
  verbs: ["create", "get", "list", "patch", "update", "watch"]
- apiGroups: ["${API_GROUP}"]
  resources: ["${CRD_PLURAL}/status"]
  verbs: ["get", "patch", "update"]
EOF

echo -e "${GREEN}✅ RBAC rules generated: ${OUTPUT_FILE}${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Review the generated RBAC rules"
echo "2. Apply to cluster: kubectl apply -f ${OUTPUT_FILE}"
echo "3. Update the API user's ClusterRole to include these permissions"
echo ""
echo -e "${YELLOW}To add to existing user:${NC}"
echo "Edit scripts/create-k8s-user.sh and add the new resource permissions"