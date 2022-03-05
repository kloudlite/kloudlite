package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"kloudlite.io/apps/message-consumer/internal/app"
	"kloudlite.io/apps/message-consumer/internal/domain"
	"kloudlite.io/pkg/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeConfig(isDev bool) (cfg *rest.Config, e error) {
	defer errors.HandleErr(&e)
	if !isDev {
		cfg, e = rest.InClusterConfig()
		errors.AssertNoError(e, fmt.Errorf("failed to get kube config because %v", e))
		return
	}

	cfgPath, ok := os.LookupEnv("KUBECONFIG")
	errors.Assert(ok, fmt.Errorf("KUBECONFIG env var is not set"))

	cfg, e = clientcmd.BuildConfigFromFlags("", cfgPath)
	errors.AssertNoError(e, fmt.Errorf("failed to get kube config because %v", e))

	return
}

func connectToK8s(isDev bool) (kcli *kubernetes.Clientset, e error) {
	defer errors.HandleErr(&e)
	cfg, e := getKubeConfig(isDev)
	errors.AssertNoError(e, fmt.Errorf("failed to get kube config because %v", e))
	return kubernetes.NewForConfig(cfg)
}

func MakeKubeApplier(isDev bool) (applier *domain.K8sApplier, e error) {
	defer errors.HandleErr(&e)
	clientset, e := connectToK8s(isDev)
	errors.Assert(e == nil, fmt.Errorf("failed to connect to k8s because %v", e))

	applier = &domain.K8sApplier{
		Apply: func(body *batchv1.Job) (e error) {
			defer errors.HandleErr(&e)
			jobs := clientset.BatchV1().Jobs(body.ObjectMeta.Namespace)
			job, e := jobs.Create(context.Background(), body, metav1.CreateOptions{})

			errors.AssertNoError(e, fmt.Errorf("failed to create job because %v", e))

			watcher, e := jobs.Watch(context.Background(), metav1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.namespace=%s", job.ObjectMeta.Namespace),
			})
			errors.AssertNoError(e, fmt.Errorf("failed to watch job because %v", e))

			for {
				result := <-watcher.ResultChan()

				switch result.Type {
				case watch.Added:
					fmt.Println(watch.Added)
				case watch.Deleted:
					fmt.Println(watch.Deleted)
				case watch.Error:
					fmt.Println(watch.Error)
				case watch.Modified:
					fmt.Println(watch.Modified)
					j := result.Object.(*batchv1.Job)
					if j.Status.Succeeded > 0 {
						fmt.Println("Job completed")
						break
					}
				default:
					logrus.Error("Unknown event type: %T", result.Type)
				}
			}
		},
	}

	return
}

type gqlClientI struct{}

func MakeGqlClient() *app.GqlClient {
	return &app.GqlClient{
		Request: func(query string, variables map[string]interface{}) (req *http.Request, e error) {
			defer errors.HandleErr(&e)
			jsonBody, e := json.Marshal(map[string]interface{}{
				"query":     query,
				"variables": variables,
			})
			errors.AssertNoError(e, fmt.Errorf("failed to marshal json body because %v", e))

			req, e = http.NewRequest("POST", "http://nxt.gateway.dev.madhouselabs.io", bytes.NewBuffer(jsonBody))
			errors.AssertNoError(e, fmt.Errorf("failed to create request because %v", e))

			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("hotspot-ci", "true")

			return
		},
	}
}
