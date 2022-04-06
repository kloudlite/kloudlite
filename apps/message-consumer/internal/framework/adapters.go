package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kloudlite.io/apps/message-consumer/internal/domain"
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
				FieldSelector: fmt.Sprintf("metadata.name=%s", job.ObjectMeta.Name),
			})

			errors.AssertNoError(e, fmt.Errorf("failed to watch job because %v", e))

			for {
				result := <-watcher.ResultChan()

				switch result.Type {

				case watch.Added:
					fmt.Println("(job) ADDED")

				case watch.Deleted:
					fmt.Println("(job) DELETED")

				case watch.Modified:
					fmt.Println("(job) MODIFIED")
					j := result.Object.(*batchv1.Job)

					if j.Status.Succeeded > 0 {
						fmt.Println("(job) COMPLETED")
						return nil
					}

					if j.Status.Failed > 0 {
						fmt.Println("(job) FAILED")
						return fmt.Errorf("(job) FAILED")
					}

				default:
					fmt.Errorf("Unknown event type: %v", result.Type)
					return nil
				}
			}
		},
	}

	return
}

type gqlClientI struct{}

func MakeGqlClient(httpClient *http.Client) *domain.GqlClient {
	Request := func(query string, variables map[string]interface{}) (req *http.Request, e error) {
		defer errors.HandleErr(&e)
		jsonBody, e := json.Marshal(map[string]interface{}{
			"query":     query,
			"variables": variables,
		})
		errors.AssertNoError(e, fmt.Errorf("failed to marshal json body because %v", e))

		gatewayUrl, ok := os.LookupEnv("GATEWAY_URL")
		errors.Assert(ok, fmt.Errorf("env 'GATEWAY_URL' not found"))
		req, e = http.NewRequest("POST", gatewayUrl, bytes.NewBuffer(jsonBody))
		errors.AssertNoError(e, fmt.Errorf("failed to create request because %v", e))

		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("hotspot-ci", "true")

		return req, nil
	}

	DoRequest := func(query string, variables map[string]interface{}) (res *http.Response, respB []byte, e error) {
		defer errors.HandleErr(&e)
		req, e := Request(query, variables)
		errors.AssertNoError(e, fmt.Errorf("could not build graphql request"))
		resp, e := httpClient.Do(req)
		errors.AssertNoError(e, fmt.Errorf("failed while making graphql request"))

		respB, e = io.ReadAll(resp.Body)
		return resp, respB, e
	}

	return &domain.GqlClient{
		Request,
		DoRequest,
	}
}
