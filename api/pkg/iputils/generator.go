package iputils

import (
	"fmt"

	"github.com/seancfoley/ipaddress-go/ipaddr"
)

var ErrIPsMaxedOut = fmt.Errorf("maximum IPs allocated")

func GenIPAddr(cidr string, offset int) (string, error) {
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

func GetIPAddrInARange(cidr string, offset int, maxOffset int) (string, error) {
	deviceRange := ipaddr.NewIPAddressString(cidr)

	address, err := deviceRange.ToAddress()
	if err != nil {
		return "", err
	}

	calcIP := address.Increment(int64(offset))
	if ok := deviceRange.Contains(calcIP.ToAddressString()); !ok {
		return "", ErrIPsMaxedOut
	}

	maxxedIP := address.Increment(int64(maxOffset))

	if maxxedIP.Compare(calcIP) > 0 {
		return ipaddr.NewIPAddressString(calcIP.GetNetIP().String()).String(), nil
	}

	return "", ErrIPsMaxedOut
}
