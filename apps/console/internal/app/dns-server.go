package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/miekg/dns"
)

type dnsHandler struct {
	logger               logging.Logger
	serviceBindingDomain domain.ServiceBindingDomain
}

const (
	DefaultDNSTTL = 5
)

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	logger := h.logger.WithKV("qname", r.Question[0].Name, "qtype", r.Question[0].Qtype)
	logger.Debugf("incoming dns request")
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	// ctx, cf := context.WithTimeout(context.TODO(), 5*time.Second)
	ctx, cf := context.WithCancel(context.TODO())
	defer cf()

	for _, question := range r.Question {
		answers := h.resolver(ctx, question.Name, question.Qtype)
		if answers == nil {
			msg.Rcode = dns.RcodeNameError
			continue
		}
		msg.Answer = append(msg.Answer, answers...)
	}

	w.WriteMsg(msg)
	logger.Debugf("outgoing dns request", "answers", msg.Answer)
}

func (h *dnsHandler) newRR(domain string, ttl int, ip string) []dns.RR {
	r, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", domain, ttl, ip))
	if err != nil {
		h.logger.Errorf(err, "failed to create dns record")
		panic(err)
	}
	return []dns.RR{r}
}

func (h *dnsHandler) resolver(ctx context.Context, domain string, qtype uint16) []dns.RR {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	question := m.Question[0]
	sp := strings.SplitN(question.Name, ".devprod.sh", 2)
	if len(sp) < 2 {
		return nil
	}

	comps := strings.Split(sp[0], ".")
	accountName := comps[len(comps)-1]
	hostname := strings.Join(comps[:len(comps)-1], ".")

	sb, err := h.serviceBindingDomain.FindServiceBindingByHostname(ctx, accountName, hostname)
	if err != nil {
		return nil
	}
	if sb == nil {
		return nil
	}

	return h.newRR(question.Name, DefaultDNSTTL, sb.Spec.GlobalIP)
}
