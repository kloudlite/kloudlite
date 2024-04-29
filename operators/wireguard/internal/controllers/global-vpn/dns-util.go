package globalvpn

import (
	"fmt"
	"sort"
	"strings"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/apps/server"

	rApi "github.com/kloudlite/operator/pkg/operator"
)

func (r *Reconciler) getCustomCoreDnsConfig(req *rApi.Request[*wgv1.GlobalVPN], corednsSvcIP string) (string, error) {
	obj, _ := req.Object, req.Context()

	updatedContent := ""

	devices := []string{}

	for _, p := range obj.Spec.Peers {
		if p.ClusterName == "" && p.DeviceName == "" {
			continue
		}

		if p.DeviceName != "" {
			devices = append(devices, fmt.Sprintf("      %s %s.device.local", p.IP, p.DeviceName))
			continue
		}

		if p.ClusterName == r.Env.ClusterName {
			updatedContent += fmt.Sprintf(`
%s.local:53 {
        errors
        kubernetes %s.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
        }
        cache 30
        loop
        reload
        loadbalance
}
`, p.ClusterName, p.ClusterName)
			continue
		}

		updatedContent += fmt.Sprintf(`
%s.local:53 {
  errors

  rewrite name regex (.*)\.svc\.%s\.local {1}.svc.cluster.local answer auto

  forward . %s

  cache 30
  loop
  reload
  loadbalance
}
`, p.ClusterName, p.ClusterName, p.IP)

	}

	devRecords := strings.Join(devices, "\n")
	updatedContent += strings.TrimSpace(fmt.Sprintf(`
device.local {
  log
  errors
  cache 30
  loop
  reload
  hosts {
%s
  }
  any
}
`, devRecords))

	return strings.TrimSpace(updatedContent), nil
}

func (r *Reconciler) getSidecarCoreDnsConfig(req *rApi.Request[*wgv1.GlobalVPN], corednsSvcIP string) (string, error) {
	exposeServices, ok := rApi.GetLocal[map[string]server.SvcInfo](req, "expose-services")
	if !ok {
		return "", fmt.Errorf("expose-services not found")
	}

	if len(exposeServices) == 0 {
		return strings.TrimSpace(fmt.Sprintf(`
.:53 {
  log
  errors

  forward . %s
  cache 30
  loop
  reload
  any
}`, corednsSvcIP)), nil
	}

	fr := []string{}
	for vip, svc := range exposeServices {
		fr = append(fr, fmt.Sprintf("      %s %s.%s.svc.cluster.local", vip, svc.Name, svc.Namespace))
	}

	sort.Slice(fr, func(i, j int) bool {
		return fr[i] < fr[j]
	})

	records := strings.Join(fr, "\n")

	return strings.TrimSpace(fmt.Sprintf(`
local {
  log
  errors
  cache 30
  loop
  reload
  hosts {
%s
  }
  any
}
`, records)), nil
}
