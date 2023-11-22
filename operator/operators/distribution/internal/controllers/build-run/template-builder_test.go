package buildrun

import (
	"fmt"
	"testing"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOutput(t *testing.T) {
	msg, err := getBuildTemplate(&dbv1.BuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Spec: dbv1.BuildRunSpec{
			Resource: dbv1.Resource{
				Cpu:        500,
				MemoryInMb: 1000,
			},
			CacheKeyName: func(s string) *string { return &s }("test"),
			AccountName:  "test",
			Registry: dbv1.Registry{
				Username: "sample",
				Password: "pass",
				Host:     "cr.khost.dev",
				Repo: dbv1.Repo{
					Name: "abc/nginx",
					Tags: []string{"v1"},
				},
			},
			GitRepo: dbv1.GitRepo{
				Url:    "https://github.com/abdheshnayak/demo-env.git",
				Branch: "main",
			},
			BuildOptions: &dbv1.BuildOptions{
				BuildArgs: map[string]string{
					"test-arg": "test",
				},
				BuildContexts: map[string]string{
					"test-ctx": "test",
				},
				DockerfilePath:    nil,
				DockerfileContent: nil,
				TargetPlatforms: []string{
					"linux/amd64",
				},
				ContextDir: nil,
			},
		},
	})
	fmt.Println(string(msg), err)
	// if !want.MatchString(msg) || err != nil {
	// 	t.Fatalf(`Hello("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
	// }
}
