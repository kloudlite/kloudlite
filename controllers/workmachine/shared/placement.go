package shared

import (
	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	corev1 "k8s.io/api/core/v1"
)

type WorkMachinePlacement struct {
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration
}

func WorkMachineAddOnPlacement(workMachineName string) WorkMachinePlacement {
	return WorkMachinePlacement{
		NodeSelector: map[string]string{
			"kloudlite.io/workmachine": workMachineName,
		},
		Tolerations: []corev1.Toleration{
			{Key: "kloudlite.io/workmachine", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
			{Key: "node.kubernetes.io/not-ready", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: fn.Ptr(int64(0))},
			{Key: "node.kubernetes.io/unreachable", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: fn.Ptr(int64(0))},
		},
	}
}
