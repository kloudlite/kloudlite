package wgutils

import (
	"github.com/kloudlite/api/pkg/errors"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GenerateKeyPair() (privateKey string, publicKey string, err error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return "", "", errors.NewEf(err, "while generating wireguard peer key-pair")
	}

	return key.String(), key.PublicKey().String(), nil
}
