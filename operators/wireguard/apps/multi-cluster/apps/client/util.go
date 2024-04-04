package client

import (
	"context"
	"fmt"
	"net"
	"time"
)

func customDialer(server string) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		dnsServerAddress := fmt.Sprintf("%s:53", server)
		d := net.Dialer{
			Timeout: time.Millisecond * time.Duration(10000),
		}
		return d.DialContext(ctx, "udp", dnsServerAddress)
	}
}
