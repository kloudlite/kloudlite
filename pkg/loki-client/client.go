package loki_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	fn "github.com/kloudlite/api/pkg/functions"
)

type lokiClient struct {
	url             *url.URL
	opts            ClientOpts
	clientCtx       context.Context
	cancelClientCtx context.CancelFunc
}

func parseLastTimestamp(data *LogResult) (*uint64, error) {
	var lastTimestamp uint64
	for _, result := range data.Data.Result {
		for _, values := range result.Values {
			val, err := strconv.ParseUint(values[0], 10, 64)
			if err != nil {
				return nil, err
			}
			if val > lastTimestamp {
				val, err := strconv.ParseUint(values[0], 10, 64)
				if err != nil {
					return nil, err
				}
				lastTimestamp = val
			}
		}
	}

	return &lastTimestamp, nil
}

func doRequest(req *http.Request) ([]byte, error) {
	get, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	all, err := io.ReadAll(get.Body)
	if err != nil {
		return nil, err
	}

	return all, nil
}

func (l *lokiClient) createLokiHttpRequest(filter QueryArgs) (*http.Request, error) {
	streamSelectorSplits := make([]string, 0)
	for _, label := range filter.StreamSelectors {
		streamSelectorSplits = append(streamSelectorSplits, label.Key+label.Operation+fmt.Sprintf("%q", label.Value))
	}

	query := url.Values{}
	searchKeywordStr := ""

	if filter.SearchKeyword != nil {
		searchKeywordStr = *filter.SearchKeyword
	}

	query.Set("query", fmt.Sprintf("%v%v", fmt.Sprintf("{%v}", strings.Join(streamSelectorSplits, ",")), searchKeywordStr))
	query.Set("direction", "BACKWARD")
	if filter.StartTime == nil {
		filter.StartTime = fn.New(time.Now().Add(-time.Hour * 24 * 30))
	}

	query.Set("start", fmt.Sprintf("%d", filter.StartTime.UnixNano()))
	if filter.EndTime != nil {
		query.Set("end", fmt.Sprintf("%d", filter.EndTime.UnixNano()))
	}

	if filter.LimitLength == nil {
		filter.LimitLength = fn.New(1000)
	}

	query.Set("limit", fmt.Sprintf("%d", *filter.LimitLength))

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/loki/api/v1/query_range", l.url.Host), nil)
	if err != nil {
		return nil, err
	}

	if l.opts.BasicAuth != nil {
		request.SetBasicAuth(l.opts.BasicAuth.Username, l.opts.BasicAuth.Password)
	}

	request.URL.RawQuery = query.Encode()

	return request, nil
}

func (l *lokiClient) GetLogs(args QueryArgs) ([]byte, error) {
	if l.clientCtx.Err() != nil {
		return nil, l.clientCtx.Err()
	}

	req, err := l.createLokiHttpRequest(args)
	if err != nil {
		return nil, err
	}

	if l.opts.Logger != nil {
		l.opts.Logger.Debugf("Loki Http Request URL: %s", req.URL.String())
	}

	b, err := doRequest(req)
	if err != nil {
		return nil, err
	}

	if args.PreWriteFunc != nil {
		var result LogResult
		if err := json.Unmarshal(b, &result); err != nil {
			return nil, err
		}

		b2, err := args.PreWriteFunc(&result)
		if err != nil {
			return nil, err
		}
		return b2, nil
	}

	return b, nil
}

func (l *lokiClient) TailLogs(args QueryArgs, writer io.WriteCloser) error {
	defer writer.Close()

	req, err := l.createLokiHttpRequest(args)
	if err != nil {
		return err
	}

	for {
		if l.clientCtx.Err() != nil {
			return l.clientCtx.Err()
		}

		b, err := doRequest(req)
		if err != nil {
			return err
		}

		var result LogResult
		if err := json.Unmarshal(b, &result); err != nil {
			return err
		}

		lt, err := parseLastTimestamp(&result)
		if err != nil {
			return err
		}

		if _, err := func() (int, error) {
			if args.PreWriteFunc != nil {
				b2, err := args.PreWriteFunc(&result)
				if err != nil {
					return 0, err
				}
				return writer.Write(b2)
			}
			return writer.Write(b)
		}(); err != nil {
			return err
		}

		qp := req.URL.Query()
		qp.Set("start", fmt.Sprintf("%d", (*lt)+1))
		qp.Del("limit")
		qp.Del("end")
		req.URL.RawQuery = qp.Encode()

		// sleeping cause, 5 seconds delay ain't that bad for logs
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
