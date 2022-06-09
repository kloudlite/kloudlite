package dns

import (
	"context"
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"log"
	"net"
	"strconv"
)

var domainsToAddresses map[string]string = map[string]string{
	"google.com.":       "1.2.3.4",
	"jameshfisher.com.": "104.198.14.52",
}

type handler struct{}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := domainsToAddresses[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	}
	w.WriteMsg(&msg)
}

func NewServer(port uint16) *dns.Server {
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	return srv
}

type ServerOptions interface {
	GetDNSPort() uint16
}

func Fx[T ServerOptions]() fx.Option {
	return fx.Module(
		"dns", fx.Provide(func(env T) *dns.Server {
			return NewServer(env.GetDNSPort())
		}),
		fx.Invoke(func(lifecycle fx.Lifecycle, s *dns.Server) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() error {
						if err := s.ListenAndServe(); err != nil {
							log.Fatalf("Failed to set udp listener %s\n", err.Error())
							return err
						}
						return nil
					}()
					return nil
				},
			})
		}),
	)
}

func main() {
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
