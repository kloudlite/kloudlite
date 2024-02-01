package wg_svc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Action struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func ensureAppRunning() error {
	if b := isReady(); !b {
		return startApp()
	}

	return nil
}

func isReady() bool {
	httpClient := http.Client{Timeout: 200 * time.Millisecond}
	if _, err := httpClient.Get(fmt.Sprint(serviceUrl, "/healthy")); err != nil {
		return false
	}
	return true
}

func sendCommand(action string, data string) error {
	httpClient := http.Client{Timeout: 200 * time.Millisecond}
	d, err := json.Marshal(Action{
		Action: action,
		Data:   data,
	})
	if err != nil {
		return err
	}
	if _, err := httpClient.Post(fmt.Sprint(serviceUrl, "/act"), "application/json", bytes.NewBuffer(d)); err != nil {
		return err
	}
	return nil
}
