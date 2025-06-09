package main

import (
	"context"
	"fmt"
	"log"

	"github.com/codingconcepts/env"
	"github.com/gofiber/fiber/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Env struct {
	HttpPort           uint16 `env:"HTTP_PORT" default:"3000"`
	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}

	kubeconfig := &rest.Config{
		Host: env.KubernetesApiProxy,
	}

	if env.KubernetesApiProxy == "" {
		var err error
		kubeconfig, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	}

	// Create the Kubernetes client
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}

	app := fiber.New()

	app.Get("/_healthy", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status": "healthy",
		})
	})

	app.Get("/kubernetes", func(c *fiber.Ctx) error {
		b, err := clientset.Discovery().RESTClient().Get().AbsPath("/healthz").DoRaw(context.TODO())
		if err != nil {
			return err
		}

		return c.Status(200).Send(b)
	})

	app.Get("/:ns/:svc", func(c *fiber.Ctx) error {
		ns := c.Params("ns")
		svc := c.Params("svc")

		pods, unhealthy, err := checkSvcHealth(clientset, ns, svc)
		if err != nil {
			return err
		}

		healthy := pods - unhealthy

		if healthy == 0 {
			return c.Status(500).JSON(fiber.Map{
				"status":  "unhealthy",
				"running": fmt.Sprintf("%d/%d", healthy, pods),
			})
		}

		return c.Status(200).JSON(fiber.Map{
			"status":  "healthy",
			"running": fmt.Sprintf("%d/%d", healthy, pods),
		})
	})

	app.All("/*", func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"status": "not found",
		})
	})

	if err := app.Listen(fmt.Sprintf(":%d", env.HttpPort)); err != nil {
		return err
	}

	return nil
}

func checkSvcHealth(clientset *kubernetes.Clientset, ns, svcName string) (int, int, error) {

	if svcName == "" {
		return 0, 0, fmt.Errorf("Service name must be provided")
	}
	if ns == "" {
		return 0, 0, fmt.Errorf("Namespace must be provided")
	}

	// Get the service
	svc, err := clientset.CoreV1().Services(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return 0, 0, err
	}

	// List the pods using the service selector
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: svc.Spec.Selector})
	pods, err := clientset.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return 0, 0, err
	}

	count := 0
	// Check the status of each pod
	for _, pod := range pods.Items {
		if err := checkPodHealth(&pod); err != nil {
			log.Printf("Error: %v", err)
			count++
		}
	}

	return len(pods.Items), count, nil
}

func checkPodHealth(pod *v1.Pod) error {
	ready := false
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			ready = true
			break
		}
	}

	if !ready {
		return fmt.Errorf("Pod %s is not healthy", pod.Name)
	}

	return nil
}
