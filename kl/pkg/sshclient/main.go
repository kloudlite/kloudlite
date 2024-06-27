package sshclient

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"golang.org/x/crypto/ssh"
)

type portFowarder struct {
	LocalPort  string
	RemotePort string
	SSHUser    string
	SSHHost    string
	SSHPort    string
	KeyPath    string

	sshConfig *ssh.ClientConfig
	listener  net.Listener
}

func (pf *StartCh) GetId() string {
	return fmt.Sprintf("%s->%s", pf.LocalPort, pf.RemotePort)
}

func portAvailable(port string) bool {
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// Function to handle the forwarding of the connection
func (pf *portFowarder) forward(ctx context.Context, localConn net.Conn) error {
	// Establish SSH connection

	sshAddr := net.JoinHostPort(pf.SSHHost, pf.SSHPort)
	sshConn, err := ssh.Dial("tcp", sshAddr, pf.sshConfig)
	if err != nil {
		localConn.Close()
		return fmt.Errorf("failed to dial SSH: %v [%s]", err, sshAddr)
	}

	// Establish connection to the remote address
	remoteConn, err := sshConn.Dial("tcp", net.JoinHostPort(pf.SSHHost, pf.RemotePort))
	if err != nil {
		localConn.Close()
		return fmt.Errorf("failed to dial remote address: %v", err)
	}

	// Forward data between local and remote connections
	copyConn := func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()
		io.Copy(writer, reader)
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)

	return nil
}

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	return ssh.PublicKeys(key), nil
}

func newForwarder(localPort, remotePort, sshUser, sshHost, sshPort, keyPath string) (*portFowarder, error) {

	pkFile, err := publicKeyFile(keyPath)

	if err != nil {
		return nil, functions.NewE(err)
	}

	// Setup SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			pkFile,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &portFowarder{
		LocalPort:  localPort,
		RemotePort: remotePort,
		SSHUser:    sshUser,
		SSHHost:    sshHost,
		SSHPort:    sshPort,
		KeyPath:    keyPath,
		sshConfig:  sshConfig,
	}, nil
}

func (pf *portFowarder) start(ctx context.Context) error {

	l, err := net.Listen("tcp", ":"+pf.LocalPort)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", pf.LocalPort, err)
	}

	pf.listener = l

	defer l.Close()
	defer ctx.Done()

	// log.Printf("listening on %s and forwarding to %s:%s", pf.LocalPort, pf.SSHHost, pf.SSHPort)

	// Handle incoming connections
	go func() {
		for {
			localConn, err := pf.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					// Listener closed due to context cancellation
					// log.Println("listener closed")
					return
				default:
					log.Printf("failed to accept connection: %v", err)
					continue
				}
			}

			go func() {
				if err := pf.forward(ctx, localConn); err != nil {
					fn.PrintError(err)
				}
			}()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	pf.listener.Close()
	return nil
}
