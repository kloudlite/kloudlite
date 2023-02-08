package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/codingconcepts/env"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	kldns "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/dns"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"math/rand"
	"net"
	"strings"

	"github.com/miekg/dns"
	"go.uber.org/fx"
	"kloudlite.io/apps/dns/internal/app/graph"
	"kloudlite.io/apps/dns/internal/app/graph/generated"
	"kloudlite.io/apps/dns/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain        string `env:"COOKIE_DOMAIN"`
	EdgeCnameBaseDomain string `env:"EDGE_CNAME_BASE_DOMAIN" required:"true"`
	DNSDomainNames      string `env:"DNS_DOMAIN_NAMES" required:"true"`
}

type DNSHandler struct {
	domain              domain.Domain
	EdgeCnameBaseDomain string
	dnsDomainNames      []string
}

func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	var e Env
	if err := env.Set(&e); err == nil {
		h.EdgeCnameBaseDomain = e.EdgeCnameBaseDomain
	}
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Answer = []dns.RR{}
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeNS:
			for _, name := range h.dnsDomainNames {
				rr := &dns.NS{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeNS,
						Class:  q.Qclass,
						Ttl:    60,
					},
					Ns: fmt.Sprintf("%s.", name),
				}
				msg.Answer = append(msg.Answer, rr)
			}
		case dns.TypeA:
			msg.Authoritative = true
			d := q.Name
			todo := context.TODO()
			host := strings.ToLower(d[:len(d)-1])

			if strings.HasSuffix(host, h.EdgeCnameBaseDomain) {
				splits := strings.Split(host, h.EdgeCnameBaseDomain)
				queryPart := splits[0]
				querySplits := strings.Split(queryPart, ".")
				var ips []string
				var err error
				if len(querySplits) == 4 {
					regionPart := querySplits[0]
					accountPart := querySplits[1]
					clusterPart := querySplits[2]
					ips, err = h.domain.GetNodeIps(todo, &regionPart, &accountPart, clusterPart)
					if err != nil {
						fmt.Println(err)
						continue
					}
				} else if len(querySplits) == 3 {
					accountPart := querySplits[0]
					clusterPart := querySplits[1]
					ips, err = h.domain.GetNodeIps(todo, nil, &accountPart, clusterPart)
					if err != nil {
						fmt.Println(err)
						continue
					}
				} else {
					clusterPart := querySplits[0]
					ips, err = h.domain.GetNodeIps(todo, nil, nil, clusterPart)
					if err != nil {
						fmt.Println(err)
						continue
					}
				}

				//ips, err := h.domain.GetNodeIps(todo, nil)
				if len(ips) == 0 {
					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET},
					})
				}
				for _, ip := range ips {
					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
						A:   net.ParseIP(ip),
					},
					)
				}
			} else {
				record, err := h.domain.GetRecord(todo, host)
				if err != nil || record == nil {
					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET},
					})
				} else {
					rand.Shuffle(len(record.Answers), func(i, j int) {
						record.Answers[i], record.Answers[j] = record.Answers[j], record.Answers[i]
					})

					for _, a := range record.Answers {
						msg.Answer = append(msg.Answer, &dns.A{
							Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(30)},
							A:   net.ParseIP(a),
						},
						)
					}
				}
			}
		}
	}

	err := w.WriteMsg(&msg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("HERE3", msg.Answer)
}

type ConsoleClientConnection *grpc.ClientConn
type FinanceClientConnection *grpc.ClientConn

var Module = fx.Module(
	"app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Record]("records", "rec", domain.RecordIndexes),
	repos.NewFxMongoRepo[*domain.Site]("sites", "site", domain.SiteIndexes),
	repos.NewFxMongoRepo[*domain.AccountCName]("account_cnames", "dns", domain.AccountCNameIndexes),
	repos.NewFxMongoRepo[*domain.RegionCName]("region_cnames", "dns", domain.RegionCNameIndexes),
	repos.NewFxMongoRepo[*domain.NodeIps]("node_ips", "nips", domain.NodeIpIndexes),
	cache.NewFxRepo[[]*domain.Record](),
	domain.Module,

	fx.Provide(func(conn ConsoleClientConnection) console.ConsoleClient {
		return console.NewConsoleClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(func(conn FinanceClientConnection) finance.FinanceClient {
		return finance.NewFinanceClient((*grpc.ClientConn)(conn))
	}),

	fx.Invoke(func(lifecycle fx.Lifecycle, s *dns.Server, d domain.Domain, recCache cache.Repo[[]*domain.Record], env *Env) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				s.Handler = &DNSHandler{
					domain:         d,
					dnsDomainNames: strings.Split(env.DNSDomainNames, ","),
				}
				return nil
			},
		})
	}),
	fx.Provide(fxDNSGrpcServer),
	fx.Invoke(func(server *grpc.Server, dnsServer kldns.DNSServer) {
		kldns.RegisterDNSServer(server, dnsServer)
	}),
	fx.Invoke(func(
		server *fiber.App,
		d domain.Domain,
		env *Env,
		cacheClient cache.Client,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: graph.NewResolver(d)},
		)

		server.Post("/upsert-domain", func(c *fiber.Ctx) error {
			var data struct {
				Domain   string
				ARecords []string
			}
			err := json.Unmarshal(c.Body(), &data)
			if err != nil {
				return err
			}
			err = d.UpsertARecords(c.Context(), data.Domain, data.ARecords)
			if err != nil {
				return err
			}
			c.Send([]byte("done"))
			return nil

		})

		server.Post("/upsert-node-ips", func(c *fiber.Ctx) error {
			var regionIps struct {
				RegionId  string   `json:"region"`
				Cluster   string   `json:"cluster"`
				AccountId string   `json:"account"`
				Ips       []string `json:"ips"`
			}

			err := json.Unmarshal(c.Body(), &regionIps)
			if err != nil {
				return err
			}
			done := d.UpdateNodeIPs(
				c.Context(),
				regionIps.RegionId,
				regionIps.AccountId,
				regionIps.Cluster,
				regionIps.Ips,
			)
			if !done {
				return fmt.Errorf("failed to update node ips")
			}
			c.Send([]byte("done"))
			return nil
		})

		server.Get("/get-records/:domain_name", func(c *fiber.Ctx) error {
			domainName := c.Params("domain_name")
			record, err := d.GetRecord(c.Context(), domainName)
			if err != nil {
				return err
			}
			r, err := json.Marshal(record)
			if err != nil {
				return err
			}
			c.Send(r)
			return nil
		})

		server.Get("/get-region-domain/:accountId/:regionId", func(c *fiber.Ctx) error {
			accountId := c.Params("accountId")
			regionId := c.Params("regionId")
			s, err := d.GetAccountEdgeCName(c.Context(), accountId, repos.ID(regionId))
			if err != nil {
				return err
			}
			c.Send([]byte(s))
			return nil
		})

		server.Delete("/delete-domain/:domain_name", func(c *fiber.Ctx) error {
			domainName := c.Params("domain_name")

			err := d.DeleteRecords(c.Context(), domainName)

			if err != nil {
				return err
			}

			c.Send([]byte("done"))
			return nil
		})

		httpServer.SetupGQLServer(
			server,
			schema,
			httpServer.NewSessionMiddleware[*common.AuthSession](
				cacheClient,
				constants.CookieName,
				env.CookieDomain,
				constants.CacheSessionPrefix,
			),
		)
	}),
)
