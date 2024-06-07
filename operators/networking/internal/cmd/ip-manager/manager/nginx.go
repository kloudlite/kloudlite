package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
)

type genNginxStreamArgs struct {
	FromAddr string
	ToAddr   string
}

func genNginxStreamConfig(protocol string, fromAddr string, toAddr string) string {
	protocol = strings.ToLower(protocol)
	if protocol == "tcp" {
		protocol = ""
	}
	return strings.TrimSpace(fmt.Sprintf(`
server {
  listen %s %s;
  proxy_pass %s;
}
`, fromAddr, protocol, toAddr))
}

func RegisterNginxStreamConfig(svcBinding *networkingv1.ServiceBinding) []string {
	result := make([]string, 0, len(svcBinding.Spec.Ports))
	for _, port := range svcBinding.Spec.Ports {
		addr := fmt.Sprintf("%s:%d", svcBinding.Spec.GlobalIP, port.Port)
		result = append(result, genNginxStreamConfig(strings.ToLower(string(port.Protocol)), addr, fmt.Sprintf("%s.%s.svc.cluster.local:%d", svcBinding.Spec.ServiceRef.Name, svcBinding.Spec.ServiceRef.Namespace, port.Port)))
	}

	return result
}

func (m *Manager) SyncNginxStreams() error {
	streams := make([]string, 0, len(m.svcNginxStreams))
	for _, v := range m.svcNginxStreams {
		streams = append(streams, v...)
	}

	b := strings.Join(streams, "\n")
	if err := os.WriteFile(fmt.Sprintf("%s/streams.conf", m.Env.NginxStreamsDir), []byte(b), 0o644); err != nil {
		return err
	}

	if !m.Env.IsDev {
		return m.restartNginx()
	}
	return nil
}

func (m *Manager) RegisterAndSyncNginxStreams(ctx context.Context, svcBindingName string) error {
	var svcBinding networkingv1.ServiceBinding
	if err := m.kcli.Get(ctx, fn.NN("", svcBindingName), &svcBinding); err != nil {
		return err
	}

	m.svcNginxStreams[svcBinding.Spec.GlobalIP] = RegisterNginxStreamConfig(&svcBinding)
	if err := m.RestartWireguard(); err != nil {
		return err
	}

	return m.SyncNginxStreams()
}

func (m *Manager) DeregisterAndSyncNginxStreams(ctx context.Context, svcBindingIP string) error {
	delete(m.svcNginxStreams, svcBindingIP)
	return m.SyncNginxStreams()
}

func (m *Manager) restartNginx() error {
	if m.Env.IsDev {
		m.logger.Infof("Restarting nginx")
		return nil
	}
	cmd := exec.Command("nginx", "-s", "reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
