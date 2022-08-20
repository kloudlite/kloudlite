package harbor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
)

type Config interface {
	GetHarborConfig() (username string, password string, registryUrl string)
}

type Client struct {
	args *Args
	url  *url.URL
}

func (h *Client) NewAuthzRequest(ctx context.Context, method, urlPath string, body io.Reader) (*http.Request, error) {
	nUrl := func() string {
		if strings.HasPrefix(urlPath, fmt.Sprintf("/api/%s", *h.args.HarborApiVersion)) {
			return h.url.String() + urlPath
		}
		return h.url.String() + "/api/" + *h.args.HarborApiVersion + urlPath
	}()

	req, err := http.NewRequestWithContext(ctx, method, nUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.args.HarborAdminUsername, h.args.HarborAdminPassword)
	return req, nil
}

type Args struct {
	HarborAdminUsername string
	HarborAdminPassword string
	HarborRegistryHost  string
	HarborApiVersion    *string
	WebhookAddr         string
}

func NewClient(args Args) (*Client, error) {
	if args.HarborApiVersion == nil {
		args.HarborApiVersion = fn.New("v2.0")
	}
	// u, err := url.Parse("https://" + args.HarborRegistryHost + "/api/" + *args.HarborApiVersion)
	u, err := url.Parse("https://" + args.HarborRegistryHost)
	if err != nil {
		return nil, errors.NewEf(err, "registryUrl is not a valid url")
	}
	return &Client{
		args: &args,
		url:  u,
	}, nil
}
