package app

import (
	"context"
	"fmt"
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
	"net"
)

type Env struct {
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`
}

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
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.Record]("records", "rec", domain.RecordIndexes),
	repos.NewFxMongoRepo[*domain.Site]("sites", "site", domain.SiteIndexes),
	repos.NewFxMongoRepo[*domain.Verification]("site_verifications", "svrf", domain.VerificationIndexes),
	domain.Module,
	fx.Invoke(func(lifecycle fx.Lifecycle, s *dns.Server, d domain.Domain) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				s.Handler = &DNSHandler{
					d,
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
