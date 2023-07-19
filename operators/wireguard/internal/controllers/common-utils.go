package wgctrl_utils

import (
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

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
