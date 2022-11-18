package operator

import (
	corev1 "k8s.io/api/core/v1"
)

type ContainerMessage struct {
	State     string `json:"state"`
	Pod       string `json:"pod,omitempty"`
	Container string `json:"container,omitempty"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	ExitCode  int32  `json:"exitCode,omitempty"`
}

func GetMessagesFromPods(pods ...corev1.Pod) []ContainerMessage {
	cMsgs := make([]ContainerMessage, 0, len(pods))

	for i := range pods {
		for j := range pods[i].Status.ContainerStatuses {
			st := pods[i].Status.ContainerStatuses[j]
			if st.State.Terminated != nil {
				cMsgs = append(
					cMsgs, ContainerMessage{
						Pod:       pods[i].Name,
						Container: st.Name,
						State:     "terminated",
						Reason:    st.State.Terminated.Reason,
						Message:   st.State.Terminated.Message,
						ExitCode:  st.State.Terminated.ExitCode,
					},
				)
			}
			if st.State.Waiting != nil {
				cMsgs = append(
					cMsgs, ContainerMessage{
						Pod:       pods[i].Name,
						Container: st.Name,
						State:     "waiting",
						Reason:    st.State.Waiting.Reason,
						Message:   st.State.Waiting.Message,
					},
				)
			}
		}
	}
	return cMsgs
}
