package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kloudlite/api/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	cfg := &rest.Config{
		Host: "http://localhost:8082/proxy",
	}

	log.Print(cfg.String())

	var err error
	kcli, err := k8s.NewClient(cfg, nil)
	if err != nil {
		panic(err)
	}

	var pods corev1.PodList
	if err := kcli.List(context.TODO(), &pods, &client.ListOptions{
		Namespace: "kloudlite",
	}); err != nil {
		panic(err)
	}

	for i := range pods.Items {
		fmt.Println(pods.Items[i].Name)
	}
}
