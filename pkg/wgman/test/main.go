package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {
	key, _ := wgtypes.GenerateKey()
	fmt.Println(key.String())
	fmt.Println(key.PublicKey().String())
}
