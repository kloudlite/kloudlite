package workmachine

import (
	"fmt"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	wmNetworkPolicyName = "wm-isolation"
)

// ensureNetworkPolicy creates or updates the NetworkPolicy for the workmachine namespace
// This ensures only the owner's environments and system namespaces can access pods in the workmachine namespace
func (r *WorkMachineReconciler) ensureNetworkPolicy(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wmNetworkPolicyName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, policy, func() error {
		policy.Labels = map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": obj.Name,
		}

		if !fn.IsOwner(policy, obj) {
			policy.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		policy.Spec = r.buildWorkmachineNetworkPolicySpec(obj)
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update network policy: %w", err))
	}

	return check.Passed()
}

// buildWorkmachineNetworkPolicySpec builds the NetworkPolicy spec for a workmachine namespace
func (r *WorkMachineReconciler) buildWorkmachineNetworkPolicySpec(obj *v1.WorkMachine) networkingv1.NetworkPolicySpec {
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

	// Rule 2: Allow from owner's environment namespaces
	// Environment namespaces have label kloudlite.io/workmachine-name set to the workmachine name
	ownerEnvRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kloudlite.io/workmachine-name": obj.Name,
					},
				},
			},
		},
	}
	ingressRules = append(ingressRules, ownerEnvRule)

	// Rule 3: Allow intra-namespace traffic (pods within the same workmachine namespace)
	intraNsRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{}, // Empty selector = all pods in namespace
			},
		},
	}
	ingressRules = append(ingressRules, intraNsRule)

	// Rule 4: Allow from wm-ingress-controller pods in any workmachine namespace
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

	// Rule 5: Allow from VPN clients (WireGuard network)
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
func (r *WorkMachineReconciler) cleanupNetworkPolicy(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wmNetworkPolicyName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if err := r.Delete(check.Context(), policy); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete network policy: %w", err))
		}
	}

	return check.Passed()
}
