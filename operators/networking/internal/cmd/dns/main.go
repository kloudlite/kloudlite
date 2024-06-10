package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/dns/types"
	"github.com/miekg/dns"
)

type ResolverCtx struct {
	context.Context
	logger *log.Logger
}

var (
	gatewayMap = types.NewSyncMap(make(map[string]string, 20))
	serviceMap = types.NewSyncMap(make(map[string][]dns.RR, 100))
)

const (
	cloudflareDNS = "1.1.1.1:53"
	googleDNS     = "8.8.8.8:53"
)

func exchangeWithDNS(ctx ResolverCtx, m *dns.Msg, nameserver string) []dns.RR {
	m, err := dns.ExchangeContext(ctx, m, nameserver)
	if err != nil {
		log.Error("while exchanging dns message", "err", err, "nameserver", nameserver)
		return nil
	}
	if m == nil {
		return nil
	}

	return m.Answer
}

func (h *dnsHandler) resolver(ctx ResolverCtx, domain string, qtype uint16) []dns.RR {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	result := make([]dns.RR, 0, 1)

	for i := range m.Question {
		qdomain := m.Question[i].Name

		gatewayMap.Debug()
		serviceMap.Debug()

		for _, k := range gatewayMap.Keys() {
			ctx.logger.Debug("checking for question", "domain", domain, "dns-suffix", k, "local-gateway-dns", h.localGatewayDNS)

			if len(result) > 0 {
				ctx.logger.Debug("found result", "result", result)
				break
			}

			if strings.HasSuffix(qdomain, fmt.Sprintf("%s.", k)) {
				switch k {
				case "svc.cluster.local":
					result = append(result, exchangeWithDNS(ctx, m, gatewayMap.Get(k))...)

				case h.localGatewayDNS:
					{
						serviceMap.Debug()

						if h.AnswerClusterLocalIPs {
							m.Question[i].Name = strings.ReplaceAll(qdomain, h.localGatewayDNS, "svc.cluster.local")
							rr := exchangeWithDNS(ctx, m, gatewayMap.Get("svc.cluster.local"))
							for j := range rr {
								rr[j].Header().Name = strings.ReplaceAll(rr[j].Header().Name, "svc.cluster.local", h.localGatewayDNS)
							}

							result = append(result, rr...)
							continue
						}

						result = append(result, serviceMap.Get(qdomain)...)
					}

				default:
					result = append(result, exchangeWithDNS(ctx, m, gatewayMap.Get(k))...)
				}
			}
		}
	}

	if len(result) > 0 {
		return result
	}

	return exchangeWithDNS(ctx, m, cloudflareDNS)
}

type dnsHandler struct {
	AnswerClusterLocalIPs bool
	localGatewayDNS       string
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	logger := log.With("qname", r.Question[0].Name, "qtype", r.Question[0].Qtype)
	logger.Debug("incoming dns request")
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		answers := h.resolver(ResolverCtx{Context: context.TODO(), logger: logger}, question.Name, question.Qtype)
		msg.Answer = append(msg.Answer, answers...)
	}

	w.WriteMsg(msg)
	logger.Debug("outgoing dns request", "answers", msg.Answer)
}

func httpServer(addr string) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	log.Info("starting http server", "addr", addr)
	r.Put("/gateway/{dns_suffix}/{gateway_addr}", func(w http.ResponseWriter, r *http.Request) {
		dnsSuffix, gatewayAddr := chi.URLParam(r, "dns_suffix"), chi.URLParam(r, "gateway_addr")
		gatewayMap.Set(dnsSuffix, gatewayAddr)
		log.Info("registered gateway", "dns-suffix", dnsSuffix, "gateway-addr", gatewayAddr)
		w.WriteHeader(http.StatusOK)
	})

	r.Put("/service/{svc_dns}/{svc_binding_ip}", func(w http.ResponseWriter, r *http.Request) {
		svcDNS, svcBindingIP := chi.URLParam(r, "svc_dns"), chi.URLParam(r, "svc_binding_ip")
		rr, err := dns.NewRR(fmt.Sprintf("%s. 5 IN A %s", svcDNS, svcBindingIP))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		serviceMap.Set(svcDNS+".", []dns.RR{rr})
		log.Info("registered service", "svc-dns", svcDNS, "svc-binding-ip", svcBindingIP)
		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(addr, r))
}

func dnsServer(addr string, handler *dnsHandler) {
	server := &dns.Server{
		Addr:      addr,
		Net:       "udp",
		Handler:   handler,
		UDPSize:   0xffff,
		ReusePort: true,
	}

	log.Info("starting dns server", "addr", addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start server", "err", err)
	}
}

func main() {
	var (
		isDebug         bool
		wgDNSAddr       string
		localDNSAddr    string
		localGatewayDNS string
		httpAddr        string
		dnsServers      string
	)

	flag.BoolVar(&isDebug, "debug", false, "--debug")
	flag.StringVar(&wgDNSAddr, "wg-dns-addr", ":53", "--wg-dns-addr <host>:<port>")
	flag.StringVar(&localDNSAddr, "local-dns-addr", ":54", "--local-dns-addr <host>:<port>")
	flag.StringVar(&localGatewayDNS, "local-gateway-dns", "svc.cluster.local", "--local-gateway-dns <alias>")
	flag.StringVar(&httpAddr, "http-addr", ":8080", "--http-addr <host>:<port>")
	flag.StringVar(&dnsServers, "dns-servers", "", "--dns-servers dns_suffix=ip[,dns_suffix2=ip2,dns_suffix3=ip3...]")
	flag.Parse()

	if isDebug {
		log.SetLevel(log.DebugLevel)
		log.Info("logging at DEBUG level")
	}

	for _, dnsServer := range strings.Split(dnsServers, ",") {
		s := strings.SplitN(dnsServer, "=", 2)
		if len(s) != 2 {
			continue
		}

		gatewayMap.Set(s[0], s[1])
		log.Info("registered gateway", "dns-suffix", s[0], "gateway-addr", s[1])
	}

	go dnsServer(localDNSAddr, &dnsHandler{AnswerClusterLocalIPs: true, localGatewayDNS: localGatewayDNS})
	go dnsServer(wgDNSAddr, &dnsHandler{localGatewayDNS: localGatewayDNS})
	httpServer(httpAddr)
}
