package loki_server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fWebsocket "github.com/gofiber/websocket/v2"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
)

var upgrader = websocket.Upgrader{}

type StreamSelector struct {
	Key       string
	Value     string
	Operation string
}

type LokiClient interface {
	Tail(clusterId string, streamSelectors []StreamSelector, filter *string, start, end *int64, limit *int, connection *fWebsocket.Conn) error
}

type lokiClient struct {
	url  *url.URL
	opts ClientOpts
}

type logResult struct {
	Data struct {
		Result []struct {
			Values [][]string `json:"values,omitempty"`
		} `json:"result,omitempty"`
	} `json:"data"`
}

func (l *lokiClient) Tail(clusterId string, streamSelectors []StreamSelector, filter *string, start, end *int64, limit *int, connection *fWebsocket.Conn) error {
	streamSelectorSplits := make([]string, 0)
	for _, label := range streamSelectors {
		streamSelectorSplits = append(streamSelectorSplits, label.Key+label.Operation+fmt.Sprintf("\"%s\"", label.Value))
	}
	query := url.Values{}
	filterStr := ""
	if filter != nil {
		filterStr = *filter
	}
	query.Set("query", fmt.Sprintf("%v%v", fmt.Sprintf("{%v}", strings.Join(streamSelectorSplits, ",")), filterStr))
	query.Set("direction", "BACKWARD")
	startTime := ""
	if start != nil {
		startTime = fmt.Sprintf("%v", start)
	} else {
		startTime = fmt.Sprintf("%v", time.Now().Add(-time.Hour*24*30).UnixNano())
	}
	query.Set("start", startTime)
	if end != nil {
		query.Set("end", fmt.Sprintf("%v", end))
	}
	if limit != nil {
		query.Set("limit", fmt.Sprintf("%v", limit))
	} else {
		query.Set("limit", fmt.Sprintf("%v", 1000))
	}
	for {
		request, err := http.NewRequest(
			http.MethodGet, fmt.Sprintf("https://%s/loki/api/v1/query_range", strings.Replace(l.url.String(), "REPLACE_ME", clusterId, 1)),
			nil,
		)
		if err != nil {
			return err
		}
		if l.opts.BasicAuth != nil {
			username := func() string {
				if l.opts.BasicAuth.Username == "" {
					return clusterId
				}
				return l.opts.BasicAuth.Username
			}()
			request.SetBasicAuth(username, l.opts.BasicAuth.Password)
		}
		request.URL.RawQuery = query.Encode()

		// u := url.URL{Scheme: "http", Host: l.url.Host, Path: "/loki/api/v1/query_range", RawQuery: query.Encode()}
		// get, err := http.Get(u.String())
		get, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		all, _ := ioutil.ReadAll(get.Body)
		var data struct {
			Data struct {
				Result []struct {
					Values [][]string `json:"values,omitempty"`
				}
			}
		}
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

func NewLokiClient(serverUrl string, opts ClientOpts) (LokiClient, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return nil, err
	}
	return &lokiClient{
		url:  u,
		opts: opts,
	}, nil
}

type LogServer *fiber.App

type LokiClientOptions interface {
	GetLokiServerUrlAndOptions() (string, ClientOpts)
	GetLogServerPort() uint64
}

func NewLogServerFx[T LokiClientOptions]() fx.Option {
	return fx.Module(
		"loki-client",
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
