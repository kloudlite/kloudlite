package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kloudlite/operator/apps/multi-cluster/apps/client/env"
	"github.com/kloudlite/operator/apps/multi-cluster/apps/common"
	"github.com/kloudlite/operator/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/pkg/logging"
)

type client struct {
	env    *env.Env
	logger logging.Logger

	client     wg.Client
	privateKey []byte
	publicKey  []byte
}

var prevConf string

func (c *client) start() error {
	for {
		if err := c.reconcile(); err != nil {
			c.logger.Error(err)
			time.Sleep(time.Second * 5)
			continue
		}
		common.ReconWait()
	}
}

func (c *client) reconcile() error {
	// c.logger.Infof("reconciling start")
	// defer c.logger.Infof("reconcilation end")

	b, err := c.sendPing()
	if err != nil {
		return err
	}

	pr := common.PeerResp{}

	if err := pr.ParseJson(b); err != nil {
		return err
	}

	endpointIp, err := c.getDnsIp(pr.Endpoint)
	if err != nil {
		return err
	}

	config := Config{
		PrivateKey: string(c.privateKey),
		IpAddress:  pr.IpAddress,
		Peers: []common.Peer{
			{
				PublicKey:  pr.PublicKey,
				AllowedIPs: pr.AllowedIPs,
				Endpoint:   *endpointIp,
			},
		},
	}

	curr := config.String()
	if prevConf == curr {
		// c.logger.Infof("no change in config")
		return nil
	}

	prevConf = curr

	wgConfg, err := config.toConfigBytes()
	if err != nil {
		return err
	}

	if err := c.client.Sync(wgConfg); err != nil {
		return err
	}

	return nil
}

func (c *client) getDnsIp(domain string) (*string, error) {
	sp := strings.Split(domain, ":")

	if len(sp) != 2 {
		return nil, fmt.Errorf("Invalid endpoint: %s", domain)
	}

	dom, port := sp[0], sp[1]

	resolver := &net.Resolver{
		PreferGo: true,
		Dial:     customDialer(c.env.KubeDns),
	}

	ips, err := resolver.LookupIPAddr(context.Background(), dom)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, errors.New("No IP addresses found")
	}

	dip := ips[0].String() + ":" + port

	return &dip, nil
}

func (c *client) sendPing() ([]byte, error) {
	data := common.PeerReq{
		PublicKey: string(c.publicKey),
		IpAddress: c.env.MyIp,
	}

	b, err := data.ToJson()
	if err != nil {
		return nil, err
	}

	payload := strings.NewReader(string(b))

	// Define the custom DNS resolver
	customResolver := &net.Resolver{
		PreferGo: true,
		Dial:     customDialer(c.env.KubeDns),
	}

	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Use the custom DNS resolver to resolve the address
				separaterIndex := len(addr) - len(":http")
				host, port := addr[:separaterIndex], addr[separaterIndex+1:]
				ips, err := customResolver.LookupIPAddr(ctx, host)
				if err != nil || len(ips) == 0 {
					return nil, err // or: return nil, errors.New("couldn't resolve the host")
				}
				// Use the first IP address returned by the custom DNS resolver
				return net.Dial(network, net.JoinHostPort(ips[0].String(), port))
			},
		},
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/peer", c.env.ServerAddr), payload)

	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil || res.StatusCode != 200 {
		if err != nil {
			return nil, err
		}

		body, e := io.ReadAll(res.Body)
		if e != nil {
			return nil, e
		}
		return nil, errors.New(string(body))
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
