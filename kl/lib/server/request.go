package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib/common"
)

func klFetch(method string, variables map[string]any, cookie *string) ([]byte, error) {
	url := constants.SERVER_URL
	reqMethod := "POST"

	marshal, err := json.Marshal(map[string]any{
		"method": method,
		"args":   []any{variables},
	})
	if err != nil {
		return nil, err
	}

	payload := strings.NewReader(string(marshal))

	client := &http.Client{}
	req, err := http.NewRequest(reqMethod, url, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Add("authority", "klcli.kloudlite.io")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/json")
	if cookie != nil {
		req.Header.Add("cookie", *cookie)
	}

	s := common.NewSpinner()
	s.Start()
	res, err := client.Do(req)
	s.Stop()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(body))

	type RespData struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	var respData RespData
	err = json.Unmarshal(body, &respData)
	if err != nil {
		common.PrintError(errors.New(fmt.Sprintf("Somoe Issue with server:\n%s\n", string(body))))
		return nil, err
	}

	if len(respData.Errors) > 0 {
		var errorMessages []string
		for _, e := range respData.Errors {
			errorMessages = append(errorMessages, e.Message)
		}

		// fmt.Println("here")
		return nil, errors.New(fmt.Sprintf("Error:\n%s", strings.Join(errorMessages, "\n")))

	}

	return body, nil

}
