package domain

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"text/template"
)

func BuildUrl(repo, branch, pullToken string) (string, error) {
	parsedURL, err := url.Parse(repo)
	if err != nil {
		fmt.Println("Error parsing Repo URL:", err)
		return "", err
	}

	parsedURL.User = url.User(pullToken)
	parsedURL.Fragment = branch

	return parsedURL.String(), nil
}

type Obj struct {
	PullUrl string
}

func getTemplate(obj Obj) ([]byte, error) {

	tstr := `
URL={{ .PullUrl }}
	`

	tmpl, err := template.New("myTemplate").Parse(tstr)
	if err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	err = tmpl.Execute(out, obj)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (*Impl) GetBuildTemplate(
	ctx context.Context, provider, repo, branch, pullToken string,
) ([]byte, error) {

	switch provider {
	case "github":

		pullUrl, err := BuildUrl(repo, branch, pullToken)
		if err != nil {
			return nil, err
		}

		var obj = Obj{
			PullUrl: pullUrl,
		}

		b, err2 := getTemplate(obj)
		if err2 != nil {
			fmt.Println("Error getting template:", err2)
			return nil, err2
		}

		return b, nil

	case "gitlab":
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
