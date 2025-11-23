// Package main implements a ProxyGuard Client CLI client
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"codeberg.org/eduVPN/proxyguard"
)

// ClientLogger is the logger for the client
type ClientLogger struct{}

// Logf logs a message with arguments
func (cl *ClientLogger) Logf(msg string, params ...interface{}) {
	log.Printf(fmt.Sprintf("[Client] %s\n", msg), params...)
}

// Log logs a message
func (cl *ClientLogger) Log(msg string) {
	log.Printf("[Client] %s\n", msg)
}

func main() {
	fwmark := flag.Int("fwmark", -1, "[Linux only] The fwmark/SO_MARK to use on the underlying TCP socket. -1 is disable.")
	forwardport := flag.Int("forward-port", 51820, "The PORT from which the UDP traffic originates.")
	listenport := flag.Int("listen-port", 51821, "The PORT to listen for UDP traffic.")
	tcpsp := flag.Int("tcp-port", 0, "The PORT to use as the TCP source port. The default is 0, which means a port chosen by the kernel.")
	to := flag.String("to", "", "The IP:PORT to which to send the converted TCP traffic to. Specify the server endpoint which also runs ProxyGuard.")
	version := flag.Bool("version", false, "Show version information")
	pipss := flag.String("peer-ips", "", "Set the IP addresses (separated by commas) to use for the server peer such that DNS resolution does not fail due to timing issues of starting the proxy e.g. on boot, before DNS resolution is ready")
	flag.Parse()

	pips := strings.Split(*pipss, ",")
	if len(pips) == 1 && pips[0] == "" {
		pips = nil
	}

	if *version {
		fmt.Printf("proxyguard-client\n%s", proxyguard.Version())
		os.Exit(0)
	}
	if *to == "" {
		fmt.Fprintln(os.Stderr, "Invalid invocation error: Please supply the --to flag")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *fwmark != -1 && runtime.GOOS != "linux" {
		fmt.Fprintln(os.Stderr, "Invalid invocation warning: The --fwmark flag is a NO-OP when you're not using Linux. We will ignore it...")
		*fwmark = -1
	}

	cl := &ClientLogger{}
	proxyguard.UpdateLogger(cl)

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			// do nothing
		}
	}()

	client := proxyguard.Client{
		ListenPort:    *listenport,
		TCPSourcePort: *tcpsp,
		Fwmark:        *fwmark,
		PeerIPS:       pips,
		Peer:          *to,
	}
	_, err := client.Setup(ctx)
	if err != nil {
		select {
		case <-ctx.Done():
			cl.Log("exiting...")
		default:
			cl.Logf("error occurred when setting up client: %v", err)
		}
		return
	}
	defer client.Close()
	err = client.Tunnel(ctx, *forwardport)
	if err != nil {
		select {
		case <-ctx.Done():
			cl.Log("exiting...")
		default:
			cl.Logf("error occurred when setting up a client: %v", err)
		}
	}
}
