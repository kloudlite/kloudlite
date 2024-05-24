package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	appconsts "github.com/kloudlite/kl/app-consts"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/fwd"
)

type Proxy struct {
	logResponse bool
}

func NewProxy(logResponse bool) (*Proxy, error) {
	return &Proxy{
		logResponse: logResponse,
	}, nil
}
func InsideBox() bool {
	s, ok := os.LookupEnv("IN_DEV_BOX")
	if !ok {
		return false
	}

	return s == "true"
}

func GetHostIp() (string, error) {
	s, ok := os.LookupEnv("HOST_IP")

	if !ok {
		return "", fmt.Errorf("HOST_IP not set")
	}

	return s, nil
}

func (p *Proxy) MakeRequest(path string, params ...[]byte) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%d%s", appconsts.AppPort, path)

	if err := func() error {
		if !InsideBox() {
			return nil
		}

		hostIp, err := GetHostIp()
		if err != nil {
			return err
		}

		url = fmt.Sprintf("http://%s:%d%s", hostIp, appconsts.AppPort, path)
		return nil
	}(); err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(map[string]interface{}{}) // Use "interface{}" instead of "any"
	if err != nil {
		return nil, err
	}

	payload := strings.NewReader(string(marshal))
	if len(params) > 0 {
		payload = strings.NewReader(string(params[0]))
	}

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

func (p *Proxy) AddFwd(chMsg []fwd.StartCh) ([]byte, error) {
	params, err := json.Marshal(chMsg)
	if err != nil {
		return nil, err
	}

	b, err := p.MakeRequest("/add-proxy", params)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) RemoveFwd(chMsg []fwd.StartCh) ([]byte, error) {
	params, err := json.Marshal(chMsg)
	if err != nil {
		return nil, err
	}

	b, err := p.MakeRequest("/remove-proxy", params)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) RemoveAllFwd(chMsg fwd.StartCh) ([]byte, error) {
	params, err := json.Marshal(chMsg)
	if err != nil {
		return nil, err
	}

	b, err := p.MakeRequest("/remove-proxy-by-ssh", params)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) ListPorts(chMsg fwd.StartCh) ([]byte, error) {
	params, err := json.Marshal(chMsg)
	if err != nil {
		return nil, err
	}

	b, err := p.MakeRequest("/list-proxy-ports", params)
	if err != nil {
		return nil, err
	}

	return b, nil
}
