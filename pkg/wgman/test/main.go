package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {
	key, _ := wgtypes.GenerateKey()
	pub := key.PublicKey().String()
	pvt := key.String()
	fmt.Println(pub)
	fmt.Println(pvt)
}
