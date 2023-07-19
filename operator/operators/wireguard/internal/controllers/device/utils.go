package device

import (
	"fmt"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/server"
	corev1 "k8s.io/api/core/v1"
)

func getNs(obj *wgv1.Device) string {
	return fmt.Sprintf("wg-%s", obj.Spec.ServerName)
}

func parseDeviceSec(obj *corev1.Secret) (pub []byte, priv []byte, ip []byte, err error) {
	var ok bool
	pub, ok = obj.Data["public-key"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse public key from sec")
	}

	priv, ok = obj.Data["private-key"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse public key from sec")
	}

	ip, ok = obj.Data["ip"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse ip from sec")
	}

	return pub, priv, ip, nil
}

func checkPortsDiffer(target []corev1.ServicePort, source []wgv1.Port) bool {
	if len(target) != len(source) {
		return true
	}

	m := make(map[int32]int32, len(source))
	for i := range source {
		m[source[i].Port] = source[i].TargetPort
	}

	for i := range target {
		if target[i].Port != source[i].Port {
			return true
		}
	}

	return false
}

// method to check either the port exists int the config
func getPort(svce []server.ConfigService, id string) (int32, error) {
	for _, s := range svce {
		if s.Id == id {
			return s.ProxyPort, nil
		}
	}
	return 0, fmt.Errorf("proxy port not found in proxy config")
}
