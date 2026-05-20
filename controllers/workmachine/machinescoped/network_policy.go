package machinescoped

import (
	"context"
	"fmt"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	wmNetworkPolicyName = "wm-isolation"
)

// ensureNetworkPolicy creates or updates the NetworkPolicy for the workmachine namespace
// This ensures only the owner's environments, shared environments, and system namespaces can access pods in the workmachine namespace
func (r *MachineScopedReconciler) ensureNetworkPolicy(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wmNetworkPolicyName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	// Find environments shared with this workmachine's owner
	sharedEnvNamespaces := r.findSharedEnvironmentNamespaces(ctx, obj.Spec.OwnedBy)

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, policy, func() error {
		policy.Labels = map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": obj.Name,
		}

		if !fn.IsOwner(policy, obj) {
			policy.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		policy.Spec = r.buildWorkmachineNetworkPolicySpec(obj, sharedEnvNamespaces)
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionNetworkPolicyReady, fmt.Errorf("failed to create/update network policy: %w", err))
	}

	return markMachineReady(session, workmachineshared.ConditionNetworkPolicyReady, "network policy is ready")
}

// buildWorkmachineNetworkPolicySpec builds the NetworkPolicy spec for a workmachine namespace
func (r *MachineScopedReconciler) buildWorkmachineNetworkPolicySpec(obj *v1.WorkMachine, sharedEnvNamespaces []string) networkingv1.NetworkPolicySpec {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	// Rule 1: Allow from system namespaces (kube-system, kloudlite)
	systemRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"kube-system", "kloudlite"},
						},
					},
				},
			},
		},
	}
	ingressRules = append(ingressRules, systemRule)

	// Rule 2: Allow from owner's namespaces (environments + workmachine)
	// Uses owned-by label for dynamic selection
	if obj.Spec.OwnedBy != "" {
		sanitizedOwner := sanitizeForLabel(obj.Spec.OwnedBy)
		ownerRule := networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"kloudlite.io/owned-by": sanitizedOwner,
						},
					},
				},
			},
		}
		ingressRules = append(ingressRules, ownerRule)
	}

	// Rule 3: Allow from environments shared with this workmachine's owner
	// These are environments where owner is in the sharedWith list
	if len(sharedEnvNamespaces) > 0 {
		var peers []networkingv1.NetworkPolicyPeer
		for _, ns := range sharedEnvNamespaces {
			peers = append(peers, networkingv1.NetworkPolicyPeer{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": ns,
					},
				},
			})
		}
		sharedRule := networkingv1.NetworkPolicyIngressRule{From: peers}
		ingressRules = append(ingressRules, sharedRule)
	}

	// Rule 4: Allow intra-namespace traffic (pods within the same workmachine namespace)
	intraNsRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{}, // Empty selector = all pods in namespace
			},
		},
	}
	ingressRules = append(ingressRules, intraNsRule)

	// Rule 5: Allow from wm-ingress-controller pods in any workmachine namespace
	// This enables exposed ports to be accessible via other users' ingress controllers
	ingressControllerRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kloudlite.io/workmachine": "true",
					},
				},
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "wm-ingress-controller",
					},
				},
			},
		},
	}
	ingressRules = append(ingressRules, ingressControllerRule)

	// Rule 6: Allow from VPN clients (WireGuard network)
	// VPN clients connect via tunnel-server and need access to services in the workmachine namespace
	vpnRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: "10.17.0.0/24",
				},
			},
		},
	}
	ingressRules = append(ingressRules, vpnRule)

	return networkingv1.NetworkPolicySpec{
		// Apply to all pods EXCEPT tunnel-server (which needs external access via hostPort)
		PodSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "app",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"tunnel-server"},
				},
			},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
		},
		Ingress: ingressRules,
		// No Egress rules - allows all egress by default
	}
}

// cleanupNetworkPolicy removes the NetworkPolicy when workmachine is being deleted
func (r *MachineScopedReconciler) cleanupNetworkPolicy(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wmNetworkPolicyName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if err := r.Delete(ctx, policy); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionNetworkPolicyReady, fmt.Errorf("failed to delete network policy: %w", err))
		}
	}

	return markMachineReady(session, workmachineshared.ConditionNetworkPolicyReady, "network policy is cleaned up")
}

// findSharedEnvironmentNamespaces finds all environment namespaces where the given user is in the sharedWith list
// This allows the workmachine to receive traffic from environments shared with its owner
func (r *MachineScopedReconciler) findSharedEnvironmentNamespaces(ctx context.Context, owner string) []string {
	if owner == "" {
		return nil
	}

	var envList environmentsv1.EnvironmentList
	if err := r.List(ctx, &envList, client.InNamespace("")); err != nil {
		return nil
	}

	var sharedNamespaces []string
	for _, env := range envList.Items {
		// Skip environments owned by the same user (already covered by owned-by label)
		if env.Spec.OwnedBy == owner {
			continue
		}

		// Check if owner is in the sharedWith list
		for _, sharedUser := range env.Spec.SharedWith {
			if sharedUser == owner {
				if env.Spec.TargetNamespace != "" {
					sharedNamespaces = append(sharedNamespaces, env.Spec.TargetNamespace)
				}
				break
			}
		}
	}

	return sharedNamespaces
}
