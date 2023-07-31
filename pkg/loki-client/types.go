package loki_client

import (
	"context"
	"io"
	"time"

	"kloudlite.io/pkg/logging"
)

type StreamSelector struct {
	Key       string
	Value     string
	Operation string
}

type BasicAuth struct {
	Username string
	Password string
}

type ClientOpts struct {
	Logger    logging.Logger
	BasicAuth *BasicAuth
}

type PreWriteFunc func(log *LogResult) ([]byte, error)

type QueryArgs struct {
	StreamSelectors []StreamSelector
	SearchKeyword   *string
	StartTime       *time.Time
	EndTime         *time.Time
	LimitLength     *int

	PreWriteFunc PreWriteFunc
}

type LogResult struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream,omitempty"`
			Values [][]string        `json:"values,omitempty"`
		} `json:"result,omitempty"`
	} `json:"data"`
}

type LokiClient interface {
	Ping(ctx context.Context) error
	GetLogs(args QueryArgs) ([]byte, error)
	TailLogs(args QueryArgs, writer io.WriteCloser) error
	Close()
}
