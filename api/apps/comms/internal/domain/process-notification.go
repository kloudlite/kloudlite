package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/types"
)

type notificationProcessor interface {
	Send() error
}

type npClient struct {
	domain *Impl
	ctx    context.Context
	n      *types.Notification
	nc     *entities.NotificationConf
}

func newNotificationProcessor(ctx context.Context, domain *Impl, n *types.Notification, nc *entities.NotificationConf) notificationProcessor {
	return &npClient{
		domain: domain,
		ctx:    ctx,
		n:      n,
		nc:     nc,
	}
}

func (c *npClient) Send() error {
	logger := c.domain.logger

	if err := c.handleTelegram(); err != nil {
		logger.Error(err.Error(), "failed to send telegram notification")
	}

	if err := c.handleSlack(); err != nil {
		logger.Error(err.Error(), "failed to send slack notification")
	}

	if err := c.handleWebhook(); err != nil {
		logger.Error(err.Error(), "failed to send webhook notification")
	}

	if err := c.handleConsoleUpdate(); err != nil {
		logger.Error(err.Error(), "failed to send console update notification")
	}

	return nil
}

func (c *npClient) handleConsoleUpdate() error {
	// TODO: (@abdheshnayak) - needs to be implemented
	c.domain.logger.Warn("console update notification is not implemented")

	return nil
}

func (c *npClient) handleSlack() error {
	if c.nc.Slack == nil || !c.nc.Slack.Enabled {
		return nil
	}

	m := &struct {
		Text string `json:"text"`
	}{c.n.ToPlain()}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return c.HttpsPost(c.nc.Slack.Url, b)
}

func (c *npClient) handleTelegram() error {
	if c.nc.Telegram == nil || !c.nc.Telegram.Enabled {
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.nc.Telegram.Token)

	m := &struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{c.nc.Telegram.ChatID, c.n.ToPlain()}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return c.HttpsPost(url, []byte(b))
}

func (c *npClient) handleWebhook() error {
	if c.nc.Webhook == nil || !c.nc.Webhook.Enabled {
		return nil
	}

	return c.HttpsPost(c.nc.Webhook.URL, []byte(c.n.ToPlain()))
}

func (c *npClient) HttpsPost(url string, body []byte) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
