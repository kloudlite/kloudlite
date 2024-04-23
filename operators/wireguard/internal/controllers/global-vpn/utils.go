package globalvpn

// import (
// 	"context"
// 	"fmt"
// 	"strconv"
//
// 	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
// 	fn "github.com/kloudlite/operator/pkg/functions"
// 	rApi "github.com/kloudlite/operator/pkg/operator"
// 	corev1 "k8s.io/api/core/v1"
// )

// type Sec struct {
// 	PublicKey  string `json:"publicKey"`
// 	PrivateKey string `json:"privateKey"`
// 	Id         int    `json:"id"`
// 	Port       int    `json:"port"`
// 	IpAddr     string `json:"ipAddr"`
// 	Interface  string `json:"interface"`
// 	DnsServer  string `json:"dnsServer"`
// }

// func parseVpnSec(s *corev1.Secret) (*Sec, error) {
// 	if s == nil {
// 		return nil, fmt.Errorf("secret is nil")
// 	}
//
// 	publicKey := string(s.Data["public-key"])
// 	privateKey := string(s.Data["private-key"])
//
// 	id, err := strconv.ParseInt(string(s.Data["id"]), 10, 64)
// 	if err != nil {
// 		id = 0
// 	}
//
// 	port, err := strconv.ParseInt(string(s.Data["port"]), 10, 64)
// 	if err != nil {
// 		port = 0
// 	}
//
// 	ipAddr := string(s.Data["ip-addr"])
//
// 	iface := s.Data["interface"]
//
// 	dnsServer := string(s.Data["dns-server"])
//
// 	return &Sec{
// 		PublicKey:  publicKey,
// 		PrivateKey: privateKey,
// 		Id:         int(id),
// 		Port:       int(port),
// 		IpAddr:     ipAddr,
// 		Interface:  string(iface),
// 		DnsServer:  dnsServer,
// 	}, nil
// }

// func (r *Reconciler) getVpnSec(ctx context.Context, obj *wgv1.GlobalVPN) (*Sec, error) {
// 	secName := fmt.Sprintf("%s-gateway-configs", obj.Name)
// 	sec, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, secName), &corev1.Secret{})
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return parseVpnSec(sec)
// }
