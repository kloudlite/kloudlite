package app

import (
	"context"
	"errors"
	"io"

	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/logging"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func ListPods(ctx context.Context, kcli k8s.Client, labels map[string]string) ([]corev1.Pod, error) {
	var pods corev1.PodList

	if err := kcli.List(ctx, &pods, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(labels),
	}); err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func StreamLogs(ctx context.Context, kcli k8s.Client, podsList []corev1.Pod, writer io.WriteCloser, logger logging.Logger) error {
	g, ctx := errgroup.WithContext(ctx)

	for i := range podsList {
		pod := podsList[i]
		for j := range pod.Spec.Containers {
			container := pod.Spec.Containers[j]
			g.Go(func() error {
				defer func() {
					logger.Infof("disconnected for (%s/%s/%s)", pod.Namespace, pod.Name, container.Name)
				}()
				logger.Infof("streaming logs for (%s/%s/%s)", pod.Namespace, pod.Name, container.Name)
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
