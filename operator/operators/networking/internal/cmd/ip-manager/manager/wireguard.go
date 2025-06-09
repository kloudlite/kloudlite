package manager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
)

func genGatewayWgPodPeer(podbinding *networkingv1.PodBinding) string {
	b := new(bytes.Buffer)
	podComment := fmt.Sprintf("\n# podbinding: %s/%s", podbinding.GetNamespace(), podbinding.GetName())
	if podbinding.Spec.PodRef != nil {
		podComment += fmt.Sprintf("\n# pod: %s/%s", podbinding.Spec.PodRef.Namespace, podbinding.Spec.PodRef.Name)
	}

	fmt.Fprintf(b, `
[Peer]%s
PublicKey = %s
AllowedIPs = %s/32
`, podComment, podbinding.Spec.WgPublicKey, podbinding.Spec.GlobalIP)

	if podbinding.Spec.PodIP != nil {
		fmt.Fprintf(b, strings.TrimSpace(`
Endpoint = %s:51820
PersistentKeepalive = 25
`), *podbinding.Spec.PodIP)
	}

	return strings.TrimSpace(b.String())
}

func (m *Manager) genGatewayWGConfig() string {
	postUps := make([]string, 0, len(m.svcNginxStreams))
	postDowns := make([]string, 0, len(m.svcNginxStreams))
	for k := range m.svcNginxStreams {
		postUps = append(postUps, fmt.Sprintf("PostUp = ip -4 address add %s/32 dev wg0", k))
		postDowns = append(postDowns, fmt.Sprintf("PostDown = ip -4 address del %s/32 dev wg0", k))
	}

	postUps = append(postUps, fmt.Sprintf("PostUp = ip -4 address add %s/32 dev wg0", m.Env.GatewayInternalDNSNameserver))
	postDowns = append(postDowns, fmt.Sprintf("PostDown = ip -4 address del %s/32 dev wg0", m.Env.GatewayInternalDNSNameserver))
	return fmt.Sprintf(`[Interface]
Address = %s
ListenPort = 51820
PrivateKey = %s
%s
%s

%s
%s
`, m.Env.GatewayGlobalIP,
		m.Env.GatewayWGPrivateKey,
		strings.Join(postUps, "\n"),
		strings.Join(postDowns, "\n"),
		strings.Join(fn.MapValues(m.podPeers), "\n\n"),
		m.gatewayWgExtraPeers,
	)
}

func (m *Manager) RestartWireguard() error {
	t := TimerStart(m.logger, "generating wg config")
	cfg := m.genGatewayWGConfig()
	if err := os.WriteFile(fmt.Sprintf("%s/wg0.conf", m.Env.WireguardConfigDir), []byte(cfg), 0o644); err != nil {
		return err
	}
	t.Stop()

	if m.Env.IsDev {
		m.logger.Debug("Restarting Wireguard")
		return nil
	}

	t.Reset("wg-quick down wg0")
	cmd := exec.Command("wg-quick", "down", "wg0")
	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Stderr = b
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 { // 1 is the error code for "interface not found"
				m.logger.Error(b.String(), "while", "running wg-quick down wg0")
				return err
			}
		}
	}
	t.Stop()

	t.Reset("wg-quick up wg0")
	cmd = exec.Command("wg-quick", "up", "wg0")
	b2 := new(bytes.Buffer)
	cmd.Stdout = b2
	cmd.Stderr = b2
	if err := cmd.Run(); err != nil {
		m.logger.Error(b.String(), "while", "running wg-quick up wg0")
		return err
	}
	t.Stop()

	m.logger.Info("wireguard restart successfull")
	return nil
}

func (m *Manager) RestartWireguardInline() error {
	cfg := m.genGatewayWGConfig()
	if err := os.WriteFile(fmt.Sprintf("%s/wg0.conf", m.Env.WireguardConfigDir), []byte(cfg), 0o644); err != nil {
		return err
	}

	if m.Env.IsDev {
		m.logger.Debug("Restarting Wireguard")
		return nil
	}

	cmd := exec.Command("sh", "-c", "wg syncconf wg0 <(wg-quick strip wg0)")
	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Stderr = b
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 { // 1 is the error code for "interface not found"
				log.Error(b.String(), "while", "running wg-quick syncconf wg0 <(wg-quick strip wg0)")
				return err
			}
		}
	}

	m.logger.Info("wireguard restart successfull")
	return nil
}

func (m *Manager) WgAddAddr(addr string) error {
	if m.Env.IsDev {
		log.Infof("executing cmd: ip -4 address add %s/32 dev wg0", addr)
		return nil
	}
	cmd := exec.Command("ip", "-4", "address", "add", fmt.Sprintf("%s/32", addr), "dev", "wg0")
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 1: // 1 is the error code for "interface not found"
			case 2: // 2 is the error code for "address already in use"
			default:
				return err
			}
		}
	}

	return nil
}

func (m *Manager) WgRemoveAddr(addr string) error {
	if m.Env.IsDev {
		log.Infof("executing cmd: ip -4 address del %s/32 dev wg0", addr)
		return nil
	}
	cmd := exec.Command("ip", "-4", "address", "del", fmt.Sprintf("%s/32", addr), "dev", "wg0")
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 1: // 1 is the error code for "interface not found"
			case 2: // 2 is the error code for "address not found"
			default:
				return err
			}
		}
	}

	return nil
}
