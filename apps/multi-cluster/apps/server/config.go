package server

import (
	"fmt"
	"os"
	"time"

	"github.com/kloudlite/operator/apps/multi-cluster/apps/common"
	"github.com/kloudlite/operator/apps/multi-cluster/constants"
	"github.com/kloudlite/operator/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/apps/multi-cluster/templates"
	"github.com/kloudlite/operator/pkg/logging"
	"sigs.k8s.io/yaml"
)

const (
	PEER_CONFIG_PATH = "/tmp/peer-config.json"
)

type PeerMap map[string]struct {
	time time.Time
	common.Peer
}

type IpMap map[int]string

var ipMap = make(IpMap)

var peerMap = make(PeerMap)
var config Config

type Config struct {
	Endpoint   string `json:"endpoint"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey,omitempty"`
	IpAddress  string `json:"ip"`

	Peers         []common.Peer `json:"peers,omitempty"`
	InternalPeers []common.Peer `json:"internal_peers,omitempty"`
}

func (s *Config) String() string {
	return fmt.Sprintf("%#v", *s)
}

func (s *Config) load(cPath string) error {

	if _, err := os.Stat(cPath); err != nil {
		return fmt.Errorf("config file not found: %s", cPath)
	}

	b, err := os.ReadFile(cPath)
	if err != nil {
		return err
	}

	if err := s.ParseYaml(b); err != nil {
		return err
	}

	pub, err := wg.GeneratePublicKey(config.PrivateKey)
	if err != nil {
		return err
	}

	s.PublicKey = string(pub)

	return nil
}

func (s Config) getAllAllowedIPs() []string {
	var ips []string
	for _, p := range s.Peers {
		ips = append(ips, p.AllowedIPs...)
	}

	ips = append(ips, s.IpAddress)

	return ips
}

func getIp(publicKey string) (string, int, error) {
	if s, ok := peerMap[publicKey]; ok {
		return s.IpAddress, s.IpId, nil
	}

	for i := constants.AgentIpRangeMin; i < constants.AgentIpRangeMax; i++ {
		if ipMap[i] == "" {
			ipMap[i] = publicKey
			b, err := wg.GetRemoteDeviceIp(int64(i))
			if err != nil {
				return "", 0, err
			}

			return string(b), i, nil
		}
	}

	return "", 0, fmt.Errorf("no available ip")
}

func (s *Config) upsertPeer(logger logging.Logger, p common.Peer) (*common.Peer, error) {

	ip, ipId, err := getIp(p.PublicKey)
	if err != nil {
		return nil, err
	}

	p.AllowedIPs = []string{
		fmt.Sprintf("%s/32", ip),
	}
	p.IpId = ipId

	p.IpAddress = ip

	defer func() {
		peerMap[p.PublicKey] = struct {
			time time.Time
			common.Peer
		}{
			time: time.Now(),
			Peer: p,
		}
	}()

	if pm, ok := peerMap[p.PublicKey]; ok {
		for i, p2 := range s.InternalPeers {
			if p2.PublicKey == pm.PublicKey {
				s.InternalPeers[i] = p
				return &p, nil
			}
		}
		return nil, fmt.Errorf("peer not found")
	}

	for i, peer := range s.InternalPeers {
		if peer.PublicKey == p.PublicKey {
			s.InternalPeers[i] = p
			return &p, nil
		}
	}

	s.InternalPeers = append(s.InternalPeers, p)

	return &p, nil
}

func (s *Config) cleanPeers() {
	for k, v := range peerMap {
		if time.Since(v.time) > constants.ExpiresIn*time.Second {
			delete(peerMap, k)
			delete(ipMap, v.IpId)

			for i, p := range s.InternalPeers {
				if p.PublicKey == k {
					s.InternalPeers = append(s.InternalPeers[:i], s.InternalPeers[i+1:]...)
					return
				}
			}
		}
	}

	return
}

func (s *Config) ToYaml() ([]byte, error) {
	return yaml.Marshal(s)
}

func (s *Config) ParseYaml(b []byte) error {
	return yaml.Unmarshal(b, s)
}

func (s *Config) toConfigBytes() ([]byte, error) {
	b, err := templates.ParseTemplate(templates.ServerConfg, *s)
	if err != nil {
		return nil, err
	}

	return b, nil
}
