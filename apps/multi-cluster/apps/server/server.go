package server

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/operator/apps/multi-cluster/apps/common"
	"github.com/kloudlite/operator/apps/multi-cluster/apps/server/env"
	"github.com/kloudlite/operator/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/pkg/logging"
)

type server struct {
	client wg.Client
	logger logging.Logger
	app    *fiber.App
	env    *env.Env
}

var (
	mu sync.Mutex
)

var prevConf string

func (s *server) sync() error {
	config.cleanPeers()

	var curr = config.String()
	if prevConf == curr {
		// s.logger.Infof("no change in config")
		return nil
	}

	// s.logger.Infof("config changed: %s vs %s", curr, prevConf)

	prevConf = config.String()

	b, err := config.toConfigBytes()
	if err != nil {
		return err
	}

	if err := s.client.Sync(b); err != nil {
		return err
	}

	return nil
}

func (s *server) Start() error {
	if err := config.load(s.env.ConfigPath); err != nil {
		return err
	}

	// go func() {
	// 	defer s.client.Stop()
	//
	// 	for {
	// 		if err := s.sync(); err != nil {
	// 			s.logger.Error(err)
	// 		}
	// 		common.ReconWait()
	// 	}
	// }()

	notFound := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	}

	s.app.Post("/peer", func(c *fiber.Ctx) error {
		mu.Lock()
		defer mu.Unlock()

		var p common.PeerReq
		if err := p.ParseJson(c.Body()); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		pr, err := config.upsertPeer(s.logger, common.Peer{
			PublicKey: p.PublicKey,
		})

		if err != nil {
			s.logger.Error(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		presp := common.PeerResp{
			IpAddress:  fmt.Sprintf("%s/32", pr.IpAddress),
			PublicKey:  config.PublicKey,
			Endpoint:   s.env.Endpoint,
			AllowedIPs: config.getAllAllowedIPs(),
		}

		b, err := presp.ToJson()
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if err := s.sync(); err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Send(b)
	})

	s.app.Get("/healthy", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	s.app.Get("/*", notFound)
	s.app.Post("/*", notFound)

	return nil
}
