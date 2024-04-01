package app

import (
	"context"
	"fmt"
	"io"

	"github.com/kloudlite/api/constants"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type LogParams struct {
	TrackingId string
}

func StreamLogs(ctx context.Context, kcli k8s.Client, writer io.WriteCloser, params LogParams) error {
	var pods corev1.PodList

	if err := kcli.List(ctx, &pods, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.ObservabilityTrackingKey: params.TrackingId,
		}),
	}); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	for i := range pods.Items {
		pod := pods.Items[i]
		for j := range pod.Spec.Containers {
			container := pod.Spec.Containers[j]
			g.Go(func() error {
				defer func() {
					fmt.Printf("\n-------disconnected for (%s/%s)---------\n", pod.Namespace, pod.Name)
				}()
				if err := kcli.ReadLogs(ctx, pod.Namespace, pod.Name, writer, &k8s.ReadLogsOptions{
					ContainerName: container.Name,
					SinceSeconds:  nil,
					TailLines:     fn.New(int64(300)),
				}); err != nil {
					fmt.Printf("ERR: %v", err)
					return err
					// return err
				}
				return nil
			})
		}
	}

	return g.Wait()
}
