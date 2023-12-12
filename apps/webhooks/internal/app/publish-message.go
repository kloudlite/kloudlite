package app

import (
	"context"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/redpanda"
	"go.uber.org/fx"
)

type PublishMessage struct {
	Message []byte
	Topic   string
	Key     string
}

type publisher struct {
	logger   logging.Logger
	msgChan  chan PublishMessage
	producer redpanda.Producer
}

func (p *publisher) publishMessage(pm PublishMessage) {
	p.msgChan <- pm
}

func (p *publisher) startPublishing() {
	for {
		pm, ok := <-p.msgChan
		if !ok {
			panic("message channel is closed")
		}
		ctx, _ := context.WithTimeout(context.TODO(), 5*time.Second)
		out, err := p.producer.Produce(ctx, pm.Topic, pm.Key, pm.Message)
		if err != nil {
			wErr := errors.NewEf(err, "could not produce message to topic %s", pm.Topic)
			p.logger.Infof(wErr.Error())
			continue
		}
		p.logger.WithKV(
			"produced.offset", out.Offset,
			"produced.topic", out.Topic,
			"produced.timestamp", out.Timestamp,
		).Infof("queued webhook")
	}
}

func PublisherFX() fx.Option {
	return fx.Module(
		"publisher",

		fx.Provide(func(logger logging.Logger, producer redpanda.Producer) *publisher {
			return &publisher{
				logger:   logger,
				msgChan:  make(chan PublishMessage),
				producer: producer,
			}
		}),

		fx.Invoke(func(lf fx.Lifecycle, p *publisher) {
			lf.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go p.startPublishing()
					return nil
				},
				OnStop: func(context.Context) error {
					close(p.msgChan)
					return nil
				},
			})
		}),
	)
}
