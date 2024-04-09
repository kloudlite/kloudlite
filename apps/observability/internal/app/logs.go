package app

import (
	"context"
	"errors"
	"io"

	"github.com/kloudlite/api/constants"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type LogParams struct {
	TrackingId string
}

func StreamLogs(ctx context.Context, kcli k8s.Client, writer io.WriteCloser, params LogParams, logger logging.Logger) error {
	var pods corev1.PodList

	if err := kcli.List(ctx, &pods, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.ObservabilityTrackingKey: params.TrackingId,
		}),
	}); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	if len(pods.Items) == 0 {
		return nil
	}

	for i := range pods.Items {
		pod := pods.Items[i]
		for j := range pod.Spec.Containers {
			container := pod.Spec.Containers[j]
			g.Go(func() error {
				defer func() {
					logger.Infof("disconnected for (%s/%s)", pod.Namespace, pod.Name)
				}()
				logger.Infof("streaming logs for (%s/%s)", pod.Namespace, pod.Name)
				if err := kcli.ReadLogs(ctx, pod.Namespace, pod.Name, writer, &k8s.ReadLogsOptions{
					ContainerName: container.Name,
					SinceSeconds:  nil,
					TailLines:     fn.New(int64(300)),
				}); err != nil {
					if !errors.Is(err, io.EOF) {
						return err
					}
					return nil
				}
				return nil
			})
		}
	}

	return g.Wait()
}
