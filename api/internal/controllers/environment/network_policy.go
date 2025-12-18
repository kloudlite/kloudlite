package environment

import (
	"context"
	"fmt"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	networkPolicyName = "env-isolation"
)

// ensureNetworkPolicy creates or updates the NetworkPolicy based on environment settings
// Network policies are enabled by default; only skip if explicitly disabled
func (r *EnvironmentReconciler) ensureNetworkPolicy(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Skip only if network policies are explicitly disabled
	if environment.Spec.NetworkPolicies != nil && !environment.Spec.NetworkPolicies.Enabled {
		// Clean up existing policy if it exists
		return r.deleteNetworkPolicy(ctx, environment, logger)
	}

	// Build the NetworkPolicy
	policy, err := r.buildNetworkPolicy(ctx, environment, logger)
	if err != nil {
		return fmt.Errorf("failed to build network policy: %w", err)
	}

	// Create or update the policy
	existing := &networkingv1.NetworkPolicy{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      networkPolicyName,
		Namespace: environment.Spec.TargetNamespace,
	}, existing)

	if apierrors.IsNotFound(err) {
		// Create new policy
		logger.Info("Creating network policy",
			zap.String("namespace", environment.Spec.TargetNamespace),
			zap.String("visibility", environment.Spec.Visibility))

		if err := r.Create(ctx, policy); err != nil {
			return fmt.Errorf("failed to create network policy: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get existing network policy: %w", err)
	} else {
		// Update existing policy
		logger.Info("Updating network policy",
			zap.String("namespace", environment.Spec.TargetNamespace),
			zap.String("visibility", environment.Spec.Visibility))

		existing.Spec = policy.Spec
		existing.Labels = policy.Labels
		if err := r.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update network policy: %w", err)
		}
	}

	// Update condition
	r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionNetworkPolicyApplied,
		metav1.ConditionTrue, "NetworkPolicyApplied", "Network policy has been applied")

	return nil
}

// buildNetworkPolicy constructs the NetworkPolicy based on visibility settings
func (r *EnvironmentReconciler) buildNetworkPolicy(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (*networkingv1.NetworkPolicy, error) {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	// Rule 1: Always allow from system namespaces (kube-system, kloudlite)
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

	// Rule 2: Visibility-based rule
	visibilityRule, err := r.buildVisibilityIngressRule(ctx, environment, logger)
	if err != nil {
		logger.Warn("Failed to build visibility ingress rule", zap.Error(err))
		// Continue without visibility rule rather than failing
	}
	if visibilityRule != nil {
		ingressRules = append(ingressRules, *visibilityRule)
	}

	// Rule 3: Allow traffic from other environments owned by the same user
	// This enables inter-environment communication (e.g., for cloning PVC data)
	ownerEnvRule := r.buildOwnerEnvironmentsIngressRule(ctx, environment, logger)
	if ownerEnvRule != nil {
		ingressRules = append(ingressRules, *ownerEnvRule)
	}

	// Rule 4: Custom ingress rules from spec (allowed namespaces and custom rules)
	customRules := r.buildCustomIngressRules(environment)
	ingressRules = append(ingressRules, customRules...)

	// Rule 5: Allow intra-namespace traffic (pods within same namespace)
	intraNsRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{}, // Empty selector = all pods in namespace
			},
		},
	}
	ingressRules = append(ingressRules, intraNsRule)

	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyName,
			Namespace: environment.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/environment": environment.Name,
				"kloudlite.io/managed":     "true",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			// Apply to all pods in namespace
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: ingressRules,
			// No Egress rules - allows all egress by default
		},
	}, nil
}

// buildVisibilityIngressRule builds ingress rule based on visibility setting
func (r *EnvironmentReconciler) buildVisibilityIngressRule(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (*networkingv1.NetworkPolicyIngressRule, error) {
	visibility := environment.Spec.Visibility
	if visibility == "" {
		visibility = "private" // default
	}

	switch visibility {
	case "private":
		// Allow only from owner's workmachine namespace
		ownerNs, err := r.getWorkMachineNamespaceForUser(ctx, environment.Spec.OwnedBy)
		if err != nil {
			logger.Warn("Could not find owner's workmachine namespace",
				zap.String("owner", environment.Spec.OwnedBy), zap.Error(err))
			return nil, nil // Skip visibility rule if owner not found
		}
		return &networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"kubernetes.io/metadata.name": ownerNs,
						},
					},
				},
			},
		}, nil

	case "shared":
		// Allow from owner + sharedWith users
		var peers []networkingv1.NetworkPolicyPeer

		// Add owner
		ownerNs, err := r.getWorkMachineNamespaceForUser(ctx, environment.Spec.OwnedBy)
		if err == nil {
			peers = append(peers, networkingv1.NetworkPolicyPeer{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": ownerNs,
					},
				},
			})
		} else {
			logger.Warn("Could not find owner's workmachine namespace",
				zap.String("owner", environment.Spec.OwnedBy), zap.Error(err))
		}

		// Add sharedWith users
		for _, username := range environment.Spec.SharedWith {
			ns, err := r.getWorkMachineNamespaceForUser(ctx, username)
			if err != nil {
				logger.Warn("Could not find workmachine namespace for shared user",
					zap.String("user", username), zap.Error(err))
				continue
			}
			peers = append(peers, networkingv1.NetworkPolicyPeer{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": ns,
					},
				},
			})
		}

		if len(peers) == 0 {
			return nil, nil
		}
		return &networkingv1.NetworkPolicyIngressRule{From: peers}, nil

	case "open":
		// Allow from all workmachine namespaces (using the label set by workmachine controller)
		return &networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"kloudlite.io/workmachine": "true",
						},
					},
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown visibility: %s", visibility)
	}
}

// buildOwnerEnvironmentsIngressRule builds an ingress rule allowing traffic from all environments owned by the same user
// This enables inter-environment communication (e.g., cloning, shared services)
func (r *EnvironmentReconciler) buildOwnerEnvironmentsIngressRule(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) *networkingv1.NetworkPolicyIngressRule {
	owner := environment.Spec.OwnedBy
	if owner == "" {
		return nil
	}

	// List all environments owned by the same user
	var envList environmentsv1.EnvironmentList
	if err := r.List(ctx, &envList, client.MatchingFields{"spec.ownedBy": owner}); err != nil {
		logger.Warn("Failed to list environments for owner", zap.String("owner", owner), zap.Error(err))
		return nil
	}

	if len(envList.Items) <= 1 {
		// Only this environment exists, no need for cross-environment rule
		return nil
	}

	var peers []networkingv1.NetworkPolicyPeer
	for _, env := range envList.Items {
		if env.Name == environment.Name {
			continue // Skip self
		}
		if env.Spec.TargetNamespace == "" {
			continue
		}
		peers = append(peers, networkingv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": env.Spec.TargetNamespace,
				},
			},
		})
	}

	if len(peers) == 0 {
		return nil
	}

	return &networkingv1.NetworkPolicyIngressRule{From: peers}
}

// getWorkMachineNamespaceForUser looks up the WorkMachine namespace for a user email
func (r *EnvironmentReconciler) getWorkMachineNamespaceForUser(ctx context.Context, userEmail string) (string, error) {
	var wmList workmachinevl.WorkMachineList
	if err := r.List(ctx, &wmList, client.MatchingFields{"spec.ownedBy": userEmail}); err != nil {
		return "", fmt.Errorf("failed to list workmachines: %w", err)
	}

	if len(wmList.Items) == 0 {
		return "", fmt.Errorf("no workmachine found for user %s", userEmail)
	}

	return wmList.Items[0].Spec.TargetNamespace, nil
}

// buildCustomIngressRules converts custom ingress rules from spec
func (r *EnvironmentReconciler) buildCustomIngressRules(environment *environmentsv1.Environment) []networkingv1.NetworkPolicyIngressRule {
	if environment.Spec.NetworkPolicies == nil {
		return nil
	}

	var rules []networkingv1.NetworkPolicyIngressRule

	// Add allowed namespaces
	if len(environment.Spec.NetworkPolicies.AllowedNamespaces) > 0 {
		var peers []networkingv1.NetworkPolicyPeer
		for _, ns := range environment.Spec.NetworkPolicies.AllowedNamespaces {
			peers = append(peers, networkingv1.NetworkPolicyPeer{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": ns,
					},
				},
			})
		}
		rules = append(rules, networkingv1.NetworkPolicyIngressRule{From: peers})
	}

	// Convert custom ingress rules
	for _, customRule := range environment.Spec.NetworkPolicies.IngressRules {
		rule := networkingv1.NetworkPolicyIngressRule{}

		// Convert From peers
		for _, peer := range customRule.From {
			npPeer := networkingv1.NetworkPolicyPeer{}
			if peer.NamespaceSelector != nil {
				npPeer.NamespaceSelector = &metav1.LabelSelector{
					MatchLabels: peer.NamespaceSelector.MatchLabels,
				}
			}
			if peer.PodSelector != nil {
				npPeer.PodSelector = &metav1.LabelSelector{
					MatchLabels: peer.PodSelector.MatchLabels,
				}
			}
			rule.From = append(rule.From, npPeer)
		}

		// Convert Ports
		for _, port := range customRule.Ports {
			protocol := corev1.Protocol(port.Protocol)
			if protocol == "" {
				protocol = corev1.ProtocolTCP
			}
			portNum := intstr.FromInt32(port.Port)
			rule.Ports = append(rule.Ports, networkingv1.NetworkPolicyPort{
				Protocol: &protocol,
				Port:     &portNum,
			})
		}

		rules = append(rules, rule)
	}

	return rules
}

// deleteNetworkPolicy removes the NetworkPolicy when disabled
func (r *EnvironmentReconciler) deleteNetworkPolicy(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	policy := &networkingv1.NetworkPolicy{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      networkPolicyName,
		Namespace: environment.Spec.TargetNamespace,
	}, policy)

	if apierrors.IsNotFound(err) {
		return nil // Already deleted
	}
	if err != nil {
		return err
	}

	logger.Info("Deleting network policy", zap.String("namespace", environment.Spec.TargetNamespace))
	if err := r.Delete(ctx, policy); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	// Update condition
	r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionNetworkPolicyApplied,
		metav1.ConditionFalse, "NetworkPolicyRemoved", "Network policy has been removed")

	return nil
}

// ensureNetworkPolicyWithOwner creates or updates the NetworkPolicy with owner reference
// Network policies are enabled by default; only skip if explicitly disabled
func (r *EnvironmentReconciler) ensureNetworkPolicyWithOwner(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Skip only if network policies are explicitly disabled
	if environment.Spec.NetworkPolicies != nil && !environment.Spec.NetworkPolicies.Enabled {
		return r.deleteNetworkPolicy(ctx, environment, logger)
	}

	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyName,
			Namespace: environment.Spec.TargetNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, policy, func() error {
		// Build the full policy spec
		builtPolicy, err := r.buildNetworkPolicy(ctx, environment, logger)
		if err != nil {
			return err
		}

		policy.Labels = builtPolicy.Labels
		policy.Spec = builtPolicy.Spec

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create or update network policy: %w", err)
	}

	r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionNetworkPolicyApplied,
		metav1.ConditionTrue, "NetworkPolicyApplied", "Network policy has been applied")

	return nil
}
