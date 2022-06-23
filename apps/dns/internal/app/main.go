package app

import (
	"context"
	"fmt"
	"net"

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
	for i, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:

			msg.Authoritative = true
			d := q.Name
			todo := context.TODO()
			host := d[:len(d)-1]

			records, err := h.domain.GetRecords(todo, host)

			if err != nil || len(records) == 0 {
				msg.Answer[i] = &dns.A{
					Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET},
				}
			}

			for _, r := range records {
				if r.Type == "A" {

					// if msg.Answer[i] == nil {
					// 	msg.Answer[i] = &dns.A{
					// 		Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: r.TTL},
					// 		A:   net.ParseIP(r.Answer),
					// 	}
					// }

					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: r.TTL},
						A:   net.ParseIP(r.Answer),
					},
					)

					fmt.Println(msg.Answer)

					// msg.Answer[i] = &dns.A{
					// 	Hdr: dns.RR_Header{Name: d, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: r.TTL},
					// 	A:   net.ParseIP(r.Answer),
					// }

				}
			}
		}
	}

	w.WriteMsg(&msg)
}

var Module = fx.Module(
	"app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Record]("records", "rec", domain.RecordIndexes),
	repos.NewFxMongoRepo[*domain.Site]("sites", "site", domain.SiteIndexes),
	repos.NewFxMongoRepo[*domain.Verification]("site_verifications", "svrf", domain.VerificationIndexes),
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
			err = d.AddARecords(c.Context(), data.Domain, data.ARecords, "kloudlite")

			if err != nil {
				return err
			}

			c.Send([]byte("done"))
			return nil

		})

		server.Delete("/delete-domain/:domain_name", func(c *fiber.Ctx) error {
			domainName := c.Params("domain_name")
			err := d.DeleteRecords(c.Context(), domainName, "kloudlite")

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
