package k8s

import (
	"testing"
	"time"

	"k8s.io/client-go/rest"
)

func TestClientConfigsKeepTimeoutOffWatchClient(t *testing.T) {
	base := &rest.Config{Host: "https://cluster.example"}

	requestConfig, watchConfig := clientConfigs(base)

	if requestConfig.Timeout != 30*time.Second {
		t.Fatalf("request timeout = %s, want 30s", requestConfig.Timeout)
	}
	if watchConfig.Timeout != 0 {
		t.Fatalf("watch timeout = %s, want 0", watchConfig.Timeout)
	}
	if base.Timeout != 0 {
		t.Fatalf("base timeout = %s, want unchanged zero value", base.Timeout)
	}
}
