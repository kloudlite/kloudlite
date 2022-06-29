package server

import (
	"encoding/json"
	"io/ioutil"
	"kloudlite.io/pkg/errors"
	"net/http"
	"strings"
)

func gql(query string, variables map[string]any, cookie *string) ([]byte, error) {
	url := "https://gateway-01.dev.kloudlite.io/"
	method := "POST"
	marshal, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}
	payload := strings.NewReader(string(marshal))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("authority", "gateway-01.dev.kloudlite.io")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/json")
	if cookie != nil {
		req.Header.Add("cookie", *cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	type RespData struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	var respData RespData
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, err
	}
	if len(respData.Errors) > 0 {
		var errorMessages []string
		for _, e := range respData.Errors {
			errorMessages = append(errorMessages, e.Message)
		}
		return nil, errors.New(strings.Join(errorMessages, "\n"))
	}
	return body, nil
}
