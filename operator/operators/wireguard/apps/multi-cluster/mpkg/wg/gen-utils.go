package wg

import (
	"fmt"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Ptr[T any](t T) *T {
	return &t
}

func GenerateWgKeys() ([]byte, []byte, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	return []byte(key.PublicKey().String()), []byte(key.String()), nil
}

func GeneratePublicKey(privateKey string) ([]byte, error) {
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return nil, err
	}

	return []byte(key.PublicKey().String()), nil
}

var ErrIPsMaxedOut error = fmt.Errorf("maximum IPs limit reached")

func GenIPAddr(offset int, cidr string) (string, error) {
	deviceRange := ipaddr.NewIPAddressString(cidr)

	address, err := deviceRange.ToAddress()
	if err != nil {
		return "", err
	}

	increment := address.Increment(int64(offset))
	if ok := deviceRange.Contains(increment.ToAddressString()); !ok {
		return "", ErrIPsMaxedOut
	}

	return ipaddr.NewIPAddressString(increment.GetNetIP().String()).String(), nil
}
