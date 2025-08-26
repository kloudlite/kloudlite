package reconciler

import (
	"encoding/json"
	"fmt"

	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Check[T]) ValidateDeploymentReady(deployment *appsv1.Deployment) StepResult {
	for _, cond := range deployment.Status.Conditions {
		switch cond.Type {
		case appsv1.DeploymentAvailable:
			if cond.Status != corev1.ConditionTrue {
				var podList corev1.PodList
				if err := c.request.client.List(
					c.Context(), &podList, &client.ListOptions{
						LabelSelector: apiLabels.SelectorFromValidatedSet(deployment.Spec.Template.Labels),
						Namespace:     deployment.Namespace,
					},
				); err != nil {
					return c.Errored(fmt.Errorf("failed to list pods: %w", err))
				}

				if len(podList.Items) > 0 {
					pMessages := fn.GetMessagesFromPods(podList.Items...)
					bMsg, err := json.Marshal(pMessages)
					if err != nil {
						return c.Errored(fmt.Errorf("failed to marshal pod message: %w", err))
					}
					return c.Errored(fmt.Errorf("deployment is not ready: %s", bMsg))
				}
			}
		case appsv1.DeploymentReplicaFailure:
			if cond.Status == corev1.ConditionTrue {
				return c.Failed(fmt.Errorf(cond.Message))
			}
		}
	}

	return c.Passed()
}
