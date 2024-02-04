package loki_server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/api/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	fWebsocket "github.com/gofiber/websocket/v2"
	"github.com/gorilla/websocket"
	fn "github.com/kloudlite/api/pkg/functions"
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
			return errors.NewE(err)
		}

		if l.opts.BasicAuth != nil {
			request.SetBasicAuth(l.opts.BasicAuth.Username, l.opts.BasicAuth.Password)
		}

		request.URL.RawQuery = query.Encode()

		get, err := http.DefaultClient.Do(request)
		if err != nil {
			return errors.NewE(err)
		}
		all, _ := io.ReadAll(get.Body)
		var data logResult
		err = json.Unmarshal(all, &data)
		if err != nil {
			return errors.NewE(err)
		}
		// fmt.Printf("DATA: %+v\n", data)
		// connection.WriteMessage(websocket.TextMessage, all)
		lastTimeStamp := query.Get("start")
		for _, result := range data.Data.Result {
			for _, values := range result.Values {
				if values[0] > lastTimeStamp {
					val, err := strconv.ParseUint(values[0], 10, 64)
					if err != nil {
						return errors.NewE(err)
					}
					lastTimeStamp = fmt.Sprintf("%v", val+1)
				}
			}
		}
		query.Set("start", lastTimeStamp)
		query.Del("limit")
		query.Del("end")
		if err = connection.WriteMessage(websocket.TextMessage, all); err != nil {
			fmt.Println("[ERROR]", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (l *lokiClient) Ping(ctx context.Context) error {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/ready", l.url.Host), nil)
	if err != nil {
		return errors.NewE(err)
	}
	r, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.NewE(err)
	}
	if r.StatusCode != http.StatusOK {
		return errors.Newf("loki server is not ready, ping check failed with status code: %d", r.StatusCode)
	}
	return nil
}

func (l *lokiClient) Close() {
	l.cancelClientCtx()
}

func NewLokiClient(httpAddr string, opts ClientOpts) (LokiClient, error) {
	u, err := url.Parse(httpAddr)
	if err != nil {
		return nil, errors.NewE(err)
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

