package app

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"kloudlite.io/apps/dns/internal/domain"
	"kloudlite.io/pkg/repos"
	"net"
)

type DNSHandler struct {
	domain domain.Domain
}

func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Answer = make([]dns.RR, len(msg.Question))
	for i, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:
			msg.Authoritative = true
			domain := q.Name
			todo := context.TODO()
			records, err := h.domain.GetRecords(todo, domain)
			if err != nil {
				fmt.Println("ERROR:", err)
				return
			}
			for _, r := range records {
				if r.Type == "A" {
					msg.Answer[i] = &dns.A{
						Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: r.TTL},
						A:   net.ParseIP(r.Answer),
					}
				}
			}
		}
	}

	w.WriteMsg(&msg)
}

var Module = fx.Module(
	"app",
	repos.NewFxMongoRepo[*domain.Record]("records", "acc", domain.RecordIndexes),
	fx.Provide(func(s *dns.Server, d domain.Domain) {
		s.Handler = &DNSHandler{
			d,
		}
	}),
)
