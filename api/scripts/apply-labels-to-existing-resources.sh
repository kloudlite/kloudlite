#!/bin/bash
set -e

# Script to apply labels to existing resources that were created before webhook changes
# This ensures label-based queries work for all resources

KUBECONFIG=${KUBECONFIG:-"/tmp/kubeconfig-remote"}
export KUBECONFIG

echo "Using KUBECONFIG: $KUBECONFIG"
echo ""

# Function to add labels to WorkMachines
add_workmachine_labels() {
    echo "=== Checking WorkMachines ==="

    # Get all WorkMachines
    MACHINES=$(kubectl get workmachines.machines.kloudlite.io -o json | jq -r '.items[] | @base64')

    if [ -z "$MACHINES" ]; then
        echo "No WorkMachines found"
        return
    fi

    for machine in $MACHINES; do
        _jq() {
            echo ${machine} | base64 --decode | jq -r ${1}
        }

        NAME=$(_jq '.metadata.name')
        OWNER=$(_jq '.spec.ownedBy')
        HAS_OWNED_BY=$(_jq '.metadata.labels."kloudlite.io/owned-by" // empty')

        if [ -z "$HAS_OWNED_BY" ]; then
            echo "  Adding labels to WorkMachine: $NAME (owner: $OWNER)"
            kubectl label workmachine "$NAME" \
                kloudlite.io/owned-by="$OWNER" \
                kloudlite.io/created-by="$OWNER" \
                --overwrite
        else
            echo "  WorkMachine $NAME already has labels"
        fi
    done
    echo ""
}

# Function to add labels to Environments
add_environment_labels() {
    echo "=== Checking Environments ==="

    # Get all Environments
    ENVS=$(kubectl get environments.environments.kloudlite.io -o json | jq -r '.items[] | @base64')

    if [ -z "$ENVS" ]; then
        echo "No Environments found"
        return
    fi

    for env in $ENVS; do
        _jq() {
            echo ${env} | base64 --decode | jq -r ${1}
        }

        NAME=$(_jq '.metadata.name')
        TARGET_NS=$(_jq '.spec.targetNamespace')
        ACTIVATED=$(_jq '.spec.activated')
        HAS_TARGET_NS_LABEL=$(_jq '.metadata.labels."kloudlite.io/target-namespace" // empty')
        HAS_ACTIVATED_LABEL=$(_jq '.metadata.labels."kloudlite.io/activated" // empty')

        LABELS_TO_ADD=""

        if [ -z "$HAS_TARGET_NS_LABEL" ] && [ -n "$TARGET_NS" ]; then
            LABELS_TO_ADD="$LABELS_TO_ADD kloudlite.io/target-namespace=$TARGET_NS"
        fi

        if [ -z "$HAS_ACTIVATED_LABEL" ]; then
            ACTIVATED_VALUE="false"
            if [ "$ACTIVATED" = "true" ]; then
                ACTIVATED_VALUE="true"
            fi
            LABELS_TO_ADD="$LABELS_TO_ADD kloudlite.io/activated=$ACTIVATED_VALUE"
        fi

        if [ -n "$LABELS_TO_ADD" ]; then
            echo "  Adding labels to Environment: $NAME"
            kubectl label environment "$NAME" $LABELS_TO_ADD --overwrite
        else
            echo "  Environment $NAME already has labels"
        fi
    done
    echo ""
}

# Function to add labels to Workspaces
add_workspace_labels() {
    echo "=== Checking Workspaces ==="

    # Get all Workspaces across all namespaces
    WORKSPACES=$(kubectl get workspaces.workspaces.kloudlite.io --all-namespaces -o json | jq -r '.items[] | @base64')

    if [ -z "$WORKSPACES" ]; then
        echo "No Workspaces found"
        return
    fi

    for workspace in $WORKSPACES; do
        _jq() {
            echo ${workspace} | base64 --decode | jq -r ${1}
        }

        NAMESPACE=$(_jq '.metadata.namespace')
        NAME=$(_jq '.metadata.name')
        OWNER=$(_jq '.spec.ownedBy')
        HAS_OWNED_BY=$(_jq '.metadata.labels."kloudlite.io/owned-by" // empty')

        if [ -z "$HAS_OWNED_BY" ]; then
            echo "  Adding labels to Workspace: $NAMESPACE/$NAME (owner: $OWNER)"
            kubectl label workspace "$NAME" -n "$NAMESPACE" \
                kloudlite.io/owned-by="$OWNER" \
                kloudlite.io/created-by="$OWNER" \
                kloudlite.io/workspace-name="$NAME" \
                --overwrite
        else
            echo "  Workspace $NAMESPACE/$NAME already has labels"
        fi
    done
    echo ""
}

# Main execution
echo "Starting label application for existing resources..."
echo ""

add_workmachine_labels
add_environment_labels
add_workspace_labels

echo "=== Summary ==="
echo "Label application completed!"
echo ""
echo "Verify with:"
echo "  kubectl get workmachines -o jsonpath='{range .items[*]}{.metadata.name}{\"\\t\"}{.metadata.labels}{\"\\n\"}{end}'"
echo "  kubectl get environments -o jsonpath='{range .items[*]}{.metadata.name}{\"\\t\"}{.metadata.labels}{\"\\n\"}{end}'"
echo "  kubectl get workspaces --all-namespaces -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}{\"\\t\"}{.metadata.labels}{\"\\n\"}{end}'"
