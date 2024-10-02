package apiclient

import (
	"context"
	"encoding/json"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func klFetch(method string, variables map[string]any, cookie *string, verbose ...bool) ([]byte, error) {
	defer spinner.Client.UpdateMessage("loading please wait")()

	url := constants.ServerURL

	marshal, err := json.Marshal(map[string]any{
		"method": method,
		"args":   []any{variables},
	})
	if err != nil {
		return nil, functions.NewE(err)
	}

	payload := strings.NewReader(string(marshal))

	// Define the custom DNS resolver
	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			// Specify the address of your custom DNS apiclient
			dnsServer := "1.1.1.1:53"
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Use the custom DNS resolver to resolve the address

				addrArray := strings.Split(addr, ":")
				host, port := addrArray[0], addrArray[1]
				ips, err := customResolver.LookupIPAddr(ctx, host)
				if err != nil || len(ips) == 0 {
					return nil, functions.NewE(err) // or: return nil, functions.Error("couldn't resolve the host")
				}
				// Use the first IP address returned by the custom DNS resolver
				return net.Dial(network, net.JoinHostPort(ips[0].String(), port))
			},
		},
	}
	req, err := http.NewRequest(http.MethodPost, url, payload)

	if err != nil {
		return nil, functions.NewE(err)
	}

	req.Header.Add("authority", "klcli.kloudlite.io")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/json")
	if cookie != nil {
		req.Header.Add("cookie", *cookie)
	}

	//f := spinner.Client.UpdateMessage("loading please wait")
	res, err := client.Do(req)
	//f()
	if err != nil || res.StatusCode != 200 {
		if err != nil {
			return nil, functions.NewE(err)
		}

		body, e := io.ReadAll(res.Body)
		if e != nil {
			return nil, e
		}
		return nil, functions.Error(string(body))
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, functions.NewE(err)
	}

	type RespData struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if len(verbose) > 0 && verbose[0] {
		fn.Println(string(body))
	}

	var respData RespData
	err = json.Unmarshal(body, &respData)
	if err != nil {
		fn.PrintError(fn.Errorf("some issue with apiclient:\n%s", string(body)))
		return nil, functions.NewE(err)
	}

	if len(respData.Errors) > 0 {
		var errorMessages []string
		for _, e := range respData.Errors {
			errorMessages = append(errorMessages, e.Message)
		}

		return nil, fn.Errorf(strings.Join(errorMessages, "\n"))
	}

	return body, nil

}

func (apic *apiClient) GetHostDNSSuffix() (string, error) {
	cookie, err := getCookie()
	if err != nil {
		return "", fn.NewE(err)
	}
	respData, err := klFetch("cli_getDNSHostSuffix", map[string]any{}, &cookie)
	if err != nil {
		return "", fn.NewE(err)
	}
	hostDNSSuffix, err := GetFromResp[string](respData)
	if err != nil {
		return "", fn.NewE(err)
	}
	return *hostDNSSuffix, nil
}
