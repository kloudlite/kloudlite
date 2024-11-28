package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	corev1 "k8s.io/api/core/v1"
)

type genNginxStreamArgs struct {
	FromAddr string
	ToAddr   string
}

func genNginxStreamConfig(svcID string, protocol string, fromAddr string, toAddr string) string {
	protocol = strings.ToLower(protocol)
	if protocol == "tcp" {
		protocol = ""
	}

	return strings.TrimSpace(fmt.Sprintf(`
upstream %s {
  server %s;
  server 127.0.0.1:80 backup;
}

server {
  listen %s %s;
  proxy_pass %s;
}
`, svcID, toAddr, fromAddr, protocol, svcID))
}

func RegisterNginxStreamConfig(svcBinding *networkingv1.ServiceBinding) []string {
	result := make([]string, 0, len(svcBinding.Spec.Ports))
	for _, port := range svcBinding.Spec.Ports {
		addr := fmt.Sprintf("%s:%d", svcBinding.Spec.GlobalIP, port.Port)

		if svcBinding.Spec.ServiceIP == nil || *svcBinding.Spec.ServiceIP == corev1.ClusterIPNone {
			host := fmt.Sprintf("%s.%s.svc.cluster.local", svcBinding.Spec.ServiceRef.Name, svcBinding.Spec.ServiceRef.Namespace)
			svcBinding.Spec.ServiceIP = &host
		}

		result = append(result,
			genNginxStreamConfig(
				fmt.Sprintf("%s_%d_%s", svcBinding.Name, port.Port, port.Protocol),
				strings.ToLower(string(port.Protocol)),
				addr,
				fmt.Sprintf("%s:%d", *svcBinding.Spec.ServiceIP, port.Port)))
	}

	return result
}

func (m *Manager) SyncNginxStreams() error {
	streams := make([]string, 0, len(m.svcNginxStreams))
	for _, v := range m.svcNginxStreams {
		streams = append(streams, v...)
	}

	b := strings.Join(streams, "\n")

	newHash := fn.Md5([]byte(b))

	if m.runningNginxStreamFileSize == len(b) && m.runningNginxStreamsMD5 == newHash {
		m.logger.Info("nginx restart request received, but stream configuration is same")
		return nil
	}

	if err := os.WriteFile(fmt.Sprintf("%s/streams.conf", m.Env.NginxStreamsDir), []byte(b), 0o644); err != nil {
		return err
	}

	if !m.Env.IsDev {
		m.runningNginxStreamsMD5 = newHash
		m.runningNginxStreamFileSize = len(b)
		return m.restartNginx()
	}
	return nil
}

// func (m *Manager) RegisterAndSyncNginxStreams(ctx context.Context, svcBindingName string) error {
// 	var svcBinding networkingv1.ServiceBinding
// 	if err := m.kcli.Get(ctx, fn.NN("", svcBindingName), &svcBinding); err != nil {
// 		return err
// 	}
//
// 	m.svcNginxStreams[svcBinding.Spec.GlobalIP] = RegisterNginxStreamConfig(&svcBinding)
// 	if err := m.RestartWireguard(); err != nil {
// 		return err
// 	}
//
// 	return m.SyncNginxStreams()
// }

func (m *Manager) DeregisterAndSyncNginxStreams(ctx context.Context, svcBindingIP string) error {
	delete(m.svcNginxStreams, svcBindingIP)
	return m.SyncNginxStreams()
}

func (m *Manager) restartNginx() error {
	if m.Env.IsDev {
		m.logger.Info("Restarting nginx")
		return nil
	}
	cmd := exec.Command("nginx", "-s", "reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
