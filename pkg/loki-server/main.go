package loki_server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	fWebsocket "github.com/gofiber/websocket/v2"
	"github.com/gorilla/websocket"
	fn "github.com/kloudlite/api/pkg/functions"
	"go.uber.org/fx"
)

var upgrader = websocket.Upgrader{}

type StreamSelector struct {
	Key       string
	Value     string
	Operation string
}

type LokiClient interface {
	Ping(ctx context.Context) error
	Tail(streamSelectors []StreamSelector, filter *string, startTime, endTime *time.Time, limitLength *int, connection *fWebsocket.Conn) error
	Close()
}

type lokiClient struct {
	url             *url.URL
	opts            ClientOpts
	clientCtx       context.Context
	cancelClientCtx context.CancelFunc
}

type logResult struct {
	Data struct {
		Result []struct {
			Values [][]string `json:"values,omitempty"`
		} `json:"result,omitempty"`
	} `json:"data"`
}

func (l *lokiClient) Tail(streamSelectors []StreamSelector, filter *string, startTime, endTime *time.Time, limitLength *int, connection *fWebsocket.Conn) error {
	streamSelectorSplits := make([]string, 0)
	for _, label := range streamSelectors {
		streamSelectorSplits = append(streamSelectorSplits, label.Key+label.Operation+fmt.Sprintf("%q", label.Value))
	}
	query := url.Values{}
	filterStr := ""
	if filter != nil {
		filterStr = *filter
	}
	query.Set("query", fmt.Sprintf("%v%v", fmt.Sprintf("{%v}", strings.Join(streamSelectorSplits, ",")), filterStr))
	query.Set("direction", "BACKWARD")
	if startTime == nil {
		startTime = fn.New(time.Now().Add(-time.Hour * 24 * 30))
	}

	query.Set("start", fmt.Sprintf("%d", startTime.UnixNano()))
	if endTime != nil {
		query.Set("end", fmt.Sprintf("%d", endTime.UnixNano()))
	}

	if limitLength == nil {
		limitLength = fn.New(1000)
	}

	query.Set("limit", fmt.Sprintf("%d", *limitLength))

	for {
		if l.clientCtx.Err() == nil {
			return l.clientCtx.Err()
		}

		request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/loki/api/v1/query_range", l.url.Host), nil)
		if err != nil {
			return err
		}

		if l.opts.BasicAuth != nil {
			request.SetBasicAuth(l.opts.BasicAuth.Username, l.opts.BasicAuth.Password)
		}

		request.URL.RawQuery = query.Encode()

		get, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		all, _ := ioutil.ReadAll(get.Body)
		var data logResult
		err = json.Unmarshal(all, &data)
		if err != nil {
			return err
		}
		// fmt.Printf("DATA: %+v\n", data)
		// connection.WriteMessage(websocket.TextMessage, all)
		lastTimeStamp := query.Get("start")
		for _, result := range data.Data.Result {
			for _, values := range result.Values {
				if values[0] > lastTimeStamp {
					val, err := strconv.ParseUint(values[0], 10, 64)
					if err != nil {
						return err
					}
					lastTimeStamp = fmt.Sprintf("%v", val+1)
				}
			}
		}
		query.Set("start", lastTimeStamp)
		query.Del("limit")
		query.Del("end")
		connection.WriteMessage(websocket.TextMessage, all)
		time.Sleep(5 * time.Second)
	}
}

func (l *lokiClient) Ping(ctx context.Context) error {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/ready", l.url.Host), nil)
	if err != nil {
		return err
	}
	r, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("loki server is not ready, ping check failed with status code: %d", r.StatusCode)
	}
	return nil
}

func (l *lokiClient) Close() {
	l.cancelClientCtx()
}

func NewLokiClient(httpAddr string, opts ClientOpts) (LokiClient, error) {
	u, err := url.Parse(httpAddr)
	if err != nil {
		return nil, err
	}

	ctx, cf := context.WithCancel(context.TODO())

	return &lokiClient{
		url:             u,
		opts:            opts,
		clientCtx:       ctx,
		cancelClientCtx: cf,
	}, nil
}

type LogServer *fiber.App

type LokiClientOptions interface {
	GetLokiServerUrlAndOptions() (string, ClientOpts)
	GetLogServerPort() uint64
}

func NewLogServerFx[T LokiClientOptions]() fx.Option {
	return fx.Module(
		"loki-server",
		fx.Provide(
			func() LogServer {
				return fiber.New()
			},
		),
		fx.Invoke(
			func(o T, app LogServer, lifecycle fx.Lifecycle) {
				var a *fiber.App
				a = app
				lifecycle.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							go a.Listen(fmt.Sprintf(":%v", o.GetLogServerPort()))
							return nil
						},
						OnStop: func(ctx context.Context) error {
							return a.Shutdown()
						},
					},
				)
			},
		),
		fx.Provide(
			func(o T) (LokiClient, error) {
				return NewLokiClient(o.GetLokiServerUrlAndOptions())
			},
		),
		fx.Invoke(
			func(app LogServer, lokiServer LokiClient) {
				var a *fiber.App
				a = app
				a.Use(
					"/", func(c *fiber.Ctx) error {
						if fWebsocket.IsWebSocketUpgrade(c) {
							c.Locals("allowed", true)
							return c.Next()
						}
						return fiber.ErrUpgradeRequired
					},
				)
			},
		),
	)
}
