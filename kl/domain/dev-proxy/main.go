package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	appconsts "github.com/kloudlite/kl/app-consts"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Proxy struct {
	verbose     bool
	logResponse bool
}

func NewProxy(verbose, logResponse bool) (*Proxy, error) {
	return &Proxy{
		verbose:     verbose,
		logResponse: logResponse,
	}, nil
}

func (p *Proxy) MakeRequest(path string) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%d%s", appconsts.AppPort, path)

	marshal, err := json.Marshal(map[string]interface{}{}) // Use "interface{}" instead of "any"
	if err != nil {
		return nil, err
	}

	payload := strings.NewReader(string(marshal))

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", res.StatusCode)
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if p.logResponse {
			fn.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *Proxy) Status() bool {
	_, err := p.MakeRequest("/healthy")
	if err != nil {
		return false
	}

	return true
}

func (p *Proxy) WgStatus() ([]byte, error) {
	b, err := p.MakeRequest("/status")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) Start() error {
	_, err := p.MakeRequest("/start")
	if err != nil {
		return err
	}

	return nil
}

func (p *Proxy) Stop() ([]byte, error) {
	b, err := p.MakeRequest("/stop")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) Restart() ([]byte, error) {
	b, err := p.MakeRequest("/restart")
	if err != nil {
		return nil, err
	}

	return b, nil
}
