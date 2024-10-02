package sshclient

import (
	"context"
	"log"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type StartCh struct {
	RemotePort string `json:"remotePort"`
	SshPort    string `json:"sshPort"`
	LocalPort  string `json:"localPort"`
}

func GetForwardController(sshUser, sshHost, keyPath string) (startCh, cancelCh chan StartCh, exitCh chan struct{}, lports map[string]StartCh, runner func()) {
	startCh = make(chan StartCh)
	cancelCh = make(chan StartCh)
	exitCh = make(chan struct{})

	lports = make(map[string]StartCh)
	return startCh, cancelCh, exitCh, lports, func() {
		ctxs := make(map[string]context.CancelFunc)

		for {
			select {
			case <-exitCh:
				fn.Println("closing all port proxy...")

				for _, sc := range lports {
					if cancel, exists := ctxs[sc.GetId()]; exists {
						cancel()
						delete(ctxs, sc.GetId())
						delete(lports, sc.LocalPort)
						fn.Printf("[-] %s\n", sc.GetId())
					}
				}

				fn.Println("closing ports done")

				return
			case i := <-startCh:

				if !portAvailable(i.LocalPort) {
					fn.Printf("port %s already in use: %s\n", i.LocalPort, lports[i.LocalPort])
					continue
				}

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				// cf = cancel
				pf, err := newForwarder(i.LocalPort, i.RemotePort, sshUser, sshHost, i.SshPort, keyPath)
				if err != nil {
					log.Println(err)
					continue
				}

				ctxs[i.GetId()] = cancel
				lports[i.LocalPort] = i
				go func() {
					if err := pf.start(ctx); err != nil {
						fn.PrintError(err)
					}
				}()
				fn.Printf("[+] %s\n", i.GetId())
			case i := <-cancelCh:
				if cancel, exists := ctxs[i.GetId()]; exists {
					cancel()
					delete(ctxs, i.GetId())
					delete(lports, i.LocalPort)
					fn.Printf("[-] %s\n", i.GetId())
				} else {
					fn.Printf("no forwarding to cancel %s\n", i.GetId())
				}
			}
		}
	}
}
