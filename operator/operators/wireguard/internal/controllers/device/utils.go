package device

import (
	"fmt"
	"sort"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
)

func parseDeviceSec(obj *corev1.Secret) (pub []byte, priv []byte, ip []byte, err error) {
	var ok bool
	pub, ok = obj.Data["device-public-key"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse public key from sec")
	}

	priv, ok = obj.Data["device-private-key"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse public key from sec")
	}

	ip, ok = obj.Data["device-ip"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse ip from sec")
	}

	return pub, priv, ip, nil
}

func parseServerSec(obj *corev1.Secret) (pub []byte, priv []byte, ip []byte, err error) {
	var ok bool
	pub, ok = obj.Data["server-public-key"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse public key from sec")
	}

	priv, ok = obj.Data["server-private-key"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse public key from sec")
	}

	ip, ok = obj.Data["server-ip"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("can't parse ip from sec")
	}

	return pub, priv, ip, nil
}

func checkPortsDiffer(target []corev1.ServicePort, source []wgv1.Port) bool {
	if len(target) != len(source) {
		return true
	}

	sort.Slice(
		target, func(i, j int) bool {
			return target[i].Port < target[j].Port
		},
	)

	sort.Slice(
		source, func(i, j int) bool {
			return source[i].Port < source[j].Port
		},
	)

	for i := range target {
		if target[i].Port != source[i].Port || int32(target[i].TargetPort.IntValue()) != source[i].TargetPort {
			return true
		}
	}

	return false
}

func GenerateWgKeys() ([]byte, []byte, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	return []byte(key.PublicKey().String()), []byte(key.String()), nil
}
func GetRemoteDeviceIp(deviceOffcet int64) ([]byte, error) {
	deviceRange := ipaddr.NewIPAddressString("10.13.0.0/16")

	if address, addressError := deviceRange.ToAddress(); addressError == nil {
		increment := address.Increment(deviceOffcet + 2)
		return []byte(ipaddr.NewIPAddressString(increment.GetNetIP().String()).String()), nil
	} else {
		return nil, addressError
	}
}
