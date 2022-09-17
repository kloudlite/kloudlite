package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"inet.af/tcpproxy"
)

func main() {
	var p tcpproxy.Proxy

	// p.AddSNIMatchRoute(
	// 	":443", func(ctx context.Context, hostname string) bool {
	// 		fmt.Println("here:....", hostname)
	// 		return strings.HasSuffix(hostname, "dev.immunify.me")
	// 	},
	// 	tcpproxy.To("ingress-lingering-darkness-controller.kl-core.svc.cluster.local:443"),
	// )

	p.AddSNIMatchRoute(
		":443", func(ctx context.Context, hostname string) bool {
			fmt.Println("here:....", hostname)
			return strings.HasSuffix(hostname, "kloudlite.io")
		},
		tcpproxy.To("ingress-purple-forest-controller.wg-kl-core.svc.cluster.local:443"),
	)

	p.AddSNIRoute(
		":443",
		"admin.dev.immunify.me",
		tcpproxy.To("ingress-lingering-darkness-controller.wg-acc-sfrbtq63sxcbn37tcl5j420uglqq.svc.cluster.local:443"),
	)

	fmt.Println("STARTING...")
	log.Fatal(p.Run())
}
