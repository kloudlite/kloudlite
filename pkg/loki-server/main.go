package loki_server

import (
	"fmt"
	fWebsocket "github.com/gofiber/websocket/v2"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"strings"
)

var upgrader = websocket.Upgrader{}

type StreamSelector struct {
	Key       string
	Value     string
	Operation string
}

type LokiClient interface {
	Tail(
		streamSelectors []StreamSelector,
		filter *string,
		start, end *int64,
		limit *int,
		connection *fWebsocket.Conn,
	) error
}

type lokiClient struct {
	url *url.URL
}

func (l *lokiClient) Tail(
	streamSelectors []StreamSelector,
	filter *string,
	start, end *int64,
	limit *int,
	connection *fWebsocket.Conn,
) error {
	streamSelectorSplits := make([]string, 0)
	for _, label := range streamSelectors {
		streamSelectorSplits = append(streamSelectorSplits, label.Key+label.Operation+label.Operation)
	}
	query := url.Values{}
	query.Set("query", fmt.Sprintf("%v%v", fmt.Sprintf("{%v}", strings.Join(streamSelectorSplits, ",")), filter))
	if start != nil {
		query.Set("start", fmt.Sprintf("%v", start))
	}
	if end != nil {
		query.Set("env", fmt.Sprintf("%v", end))
	}
	if limit != nil {
		query.Set("limit", fmt.Sprintf("%v", limit))
	}
	u := url.URL{Scheme: "ws", Host: l.url.Host, Path: "/loki/api/v1/tail", RawQuery: query.Encode()}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return err
	}
	defer conn.Close()
	for {
		msgType, message, err := conn.ReadMessage()
		connection.WriteMessage(msgType, message)
		if err != nil {
			connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, ""))
			return err
		}
		if msgType == websocket.CloseMessage {
			return nil
		}
	}
}

func NewLokiClient(serverUrl string) (LokiClient, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return nil, err
	}
	return &lokiClient{
		url: u,
	}, nil
}
