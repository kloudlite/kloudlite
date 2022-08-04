package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
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
	CookieDomain string `env:"COOKIE_DOMAIN"`
}

type DNSHandler struct {
	domain domain.Domain
}

func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Answer = []dns.RR{}
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:
			msg.Authoritative = true
			d := q.Name
			todo := context.TODO()
			host := d[:len(d)-1]
			if strings.HasSuffix(host, ".edgenet.khost.dev") {
				ips, err := h.domain.GetNodeIps(todo, nil)
				if err != nil || len(ips) == 0 {
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
			}
			records, err := h.domain.GetRecords(todo, host)
			if err != nil || len(records) == 0 {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET},
				})
			}
			for _, r := range records {
				if r.Type == "A" {
					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: r.TTL},
						A:   net.ParseIP(r.Answer),
					},
					)

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

var Module = fx.Module(
	"app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Record]("records", "rec", domain.RecordIndexes),
	repos.NewFxMongoRepo[*domain.Site]("sites", "site", domain.SiteIndexes),
	repos.NewFxMongoRepo[*domain.AccountCName]("account_cnames", "dns", domain.AccountCNameIndexes),
	repos.NewFxMongoRepo[*domain.NodeIps]("node_ips", "nips", domain.NodeIpIndexes),
	cache.NewFxRepo[[]*domain.Record](),
	domain.Module,
	fx.Invoke(func(lifecycle fx.Lifecycle, s *dns.Server, d domain.Domain, recCache cache.Repo[[]*domain.Record]) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				s.Handler = &DNSHandler{
					domain: d,
				}
				return nil
			},
		})
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
			err := c.BodyParser(&data)
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
			var ips map[string][]string
			err := c.JSON(&ips)
			if err != nil {
				return err
			}
			done := d.UpdateNodeIPs(c.Context(), ips)
			if !done {
				return fmt.Errorf("failed to update node ips")
			}
			c.Send([]byte("done"))
			return nil
		})
		server.Get("/get-records/:domain_name", func(c *fiber.Ctx) error {
			domainName := c.Params("domain_name")
			records, err := d.GetRecords(c.Context(), domainName)
			if err != nil {
				return err
			}

			r, err := json.Marshal(records)

			if err != nil {
				return err
			}

			c.Send([]byte(r))

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
				common.CookieName,
				env.CookieDomain,
				common.CacheSessionPrefix,
			),
		)
	}),
)
