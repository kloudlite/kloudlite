package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectTLSParams struct {
	// ServerName is name with which certificate is valid
	ServerName string

	// CACertPem could be optional, if host is verified through a public CA
	CACertPem []byte
}

type ConnectOpts struct {
	SecureConnect bool
	*ConnectTLSParams
}

func Connect(addr string, opts ConnectOpts) (*grpc.ClientConn, error) {
	if !opts.SecureConnect {
		return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Load system cert pool
	roots, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}

	if opts.CACertPem != nil {
		roots.AppendCertsFromPEM(opts.CACertPem)
	}

	tlsCfg := &tls.Config{RootCAs: roots}

	if opts.ServerName != "" {
		tlsCfg.ServerName = opts.ServerName
	}

	return grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
}
