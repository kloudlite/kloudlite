package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kloudlite/operator/common"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/dns/types"
	"github.com/kloudlite/operator/pkg/logging"
	"github.com/miekg/dns"
)

type ResolverCtx struct {
	context.Context
	*slog.Logger
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
		ctx.Error("while exchanging dns message", "err", err, "nameserver", nameserver)
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

		if r := serviceMap.Get(qdomain); r != nil {
			result = append(result, r...)
		}

		for _, k := range gatewayMap.Keys() {
			ctx.Debug("checking for question", "domain", domain, "dns-suffix", k, "local-gateway-dns", h.localGatewayDNS)

			if len(result) > 0 {
				ctx.Debug("found result", "result", result)
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
	logger                *slog.Logger
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	logger := h.logger.With("qname", r.Question[0].Name, "qtype", r.Question[0].Qtype)
	logger.Debug("incoming dns request")
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		answers := h.resolver(ResolverCtx{Context: context.TODO(), Logger: h.logger}, question.Name, question.Qtype)
		msg.Answer = append(msg.Answer, answers...)
	}

	w.WriteMsg(msg)
	logger.Debug("outgoing dns request", "answers", msg.Answer)
}

func httpServer(ctx Context, addr string) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	ctx.Info("starting http server", "addr", addr)
	r.Put("/gateway/{dns_suffix}/{gateway_addr}", func(w http.ResponseWriter, r *http.Request) {
		dnsSuffix, gatewayAddr := chi.URLParam(r, "dns_suffix"), chi.URLParam(r, "gateway_addr")
		gatewayMap.Set(dnsSuffix, gatewayAddr)
		ctx.Info("registered gateway", "dns-suffix", dnsSuffix, "gateway-addr", gatewayAddr)
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
		ctx.Info("registered service", "svc-dns", svcDNS, "svc-binding-ip", svcBindingIP)
		w.WriteHeader(http.StatusOK)
	})

	server := http.Server{Addr: addr, Handler: r}
	go func() {
		<-ctx.Done()
		ctx.Warn("closing http server", "addr", addr)
		server.Shutdown(context.TODO())
	}()
	server.ListenAndServe()
	ctx.Info("http server closed", "addr", addr)
}

type Context struct {
	context.Context
	*slog.Logger
}

func dnsServer(ctx Context, addr string, handler *dnsHandler) {
	server := &dns.Server{
		Addr:      addr,
		Net:       "udp",
		Handler:   handler,
		UDPSize:   0xffff,
		ReusePort: true,
	}

	go func() {
		<-ctx.Done()
		ctx.Warn("closing dns server", "addr", addr)
		server.Shutdown()
	}()

	ctx.Info("starting dns server", "addr", addr)
	if err := server.ListenAndServe(); err != nil {
		ctx.Error("failed to start DNS server, got", "err", err)
		os.Exit(1)
	}

	ctx.Info("dns server closed", "addr", addr)
}

func main() {
	start := time.Now()
	common.PrintBuildInfo()

	var (
		isDebug         bool
		wgDNSAddr       string
		localGatewayDNS string

		accountName string

		enableLocalDNS bool
		localDNSAddr   string

		enableHTTP bool
		httpAddr   string

		dnsServers   string
		serviceHosts string
	)

	flag.BoolVar(&isDebug, "debug", false, "--debug")
	flag.StringVar(&wgDNSAddr, "wg-dns-addr", ":53", "--wg-dns-addr <host>:<port>")
	flag.StringVar(&accountName, "account", "", "--account <account_name>")

	flag.BoolVar(&enableLocalDNS, "enable-local-dns", false, "--enable-local-dns")
	flag.StringVar(&localDNSAddr, "local-dns-addr", ":54", "--local-dns-addr <host>:<port>")
	flag.StringVar(&localGatewayDNS, "local-gateway-dns", "svc.cluster.local", "--local-gateway-dns <alias>")
	flag.BoolVar(&enableHTTP, "enable-http", false, "--enable-http")
	flag.StringVar(&httpAddr, "http-addr", ":8080", "--http-addr <host>:<port>")
	flag.StringVar(&dnsServers, "dns-servers", "", "--dns-servers dns_suffix=ip[,dns_suffix2=ip2,dns_suffix3=ip3...]")
	flag.StringVar(&serviceHosts, "service-hosts", "", "--service-hosts service_host=ip[,service_host2=ip2,service_host3=ip3...]")
	flag.Parse()

	logger := logging.NewSlogLogger(logging.SlogOptions{
		ShowCaller:    true,
		ShowDebugLogs: isDebug,
	})

	for _, dnsServer := range strings.Split(dnsServers, ",") {
		s := strings.SplitN(dnsServer, "=", 2)
		if len(s) != 2 {
			continue
		}

		gatewayMap.Set(s[0], s[1])
		logger.Info("registered gateway", "dns-suffix", s[0], "gateway-addr", s[1])
	}

	if accountName != "" {
		accountDNSName := "account.kloudlite.local"
		accountRecord, err := dns.NewRR(fmt.Sprintf("%s. 5 IN TXT %s", accountDNSName, accountName))
		if err != nil {
			logger.Error("failed to parse DNS Resource Record, got", "err", err)
			os.Exit(1)
		}
		serviceMap.Set(accountDNSName+".", []dns.RR{accountRecord})
	}

	for _, serviceHost := range strings.Split(serviceHosts, ",") {
		s := strings.SplitN(serviceHost, "=", 2)
		if len(s) != 2 {
			continue
		}

		rr, err := dns.NewRR(fmt.Sprintf("%s. 5 IN A %s", s[0], s[1]))
		if err != nil {
			logger.Error("failed to parse DNS Resource Record, got", "err", err)
			os.Exit(1)
		}
		serviceMap.Set(s[0]+".", []dns.RR{rr})
		logger.Info("registered service", "host", s[0], "ip", s[1])
	}

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGTERM, syscall.SIGINT)
	defer cf()

	var wg sync.WaitGroup

	if enableLocalDNS {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dnsServer(Context{Context: ctx, Logger: logger}, localDNSAddr, &dnsHandler{AnswerClusterLocalIPs: true, localGatewayDNS: localGatewayDNS, logger: logger})
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		dnsServer(Context{Context: ctx, Logger: logger}, wgDNSAddr, &dnsHandler{localGatewayDNS: localGatewayDNS, logger: logger})
	}()

	if enableHTTP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			httpServer(Context{Context: ctx, Logger: logger}, httpAddr)
		}()
	}

	common.PrintReadyBanner2(time.Since(start))
	wg.Wait()
	logger.Warn("exiting ...")
	os.Exit(0)
}
