package utils

import (
	"fmt"
	"math/rand"

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

func Nonce(size int) string {
	chars := "0123456789"
	nonceBytes := make([]byte, size)

	for i := range nonceBytes {
		nonceBytes[i] = chars[rand.Intn(len(chars))]
	}

	return string(nonceBytes)
}
