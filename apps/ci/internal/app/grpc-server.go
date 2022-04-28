package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
	"net/http"
	"net/url"
)

type server struct {
	ci.UnimplementedCIServer
	harborUsername string
	harborPassword string
	harborUrl      url.URL
	d              domain.Domain
}

func (s *server) CreatePipeline(ctx context.Context, in *ci.PipelineIn) (*ci.PipelineOutput, error) {
	i := int(in.GithubInstallationId)
	ba := make(map[string]interface{}, 0)
	if in.BuildArgs != nil {
		for k, v := range in.BuildArgs {
			ba[k] = v
		}
	}
	pipeline, err := s.d.CretePipeline(ctx, repos.ID(in.UserId), domain.Pipeline{
		Name:                 in.Name,
		ImageName:            in.ImageName,
		PipelineEnv:          in.PipelineEnv,
		GitProvider:          in.GitProvider,
		GitRepoUrl:           in.GitRepoUrl,
		DockerFile:           &in.DockerFile,
		ContextDir:           &in.ContextDir,
		GithubInstallationId: &i,
		GitlabTokenId:        in.GitlabTokenId,
		BuildArgs:            ba,
	})
	if err != nil {
		return nil, err
	}
	return &ci.PipelineOutput{PipelineId: string(pipeline.Id)}, err
}

func (s *server) checkIfProjectExists(ctx context.Context, name string) (*bool, error) {
	s.harborUrl.Query().Add("project_name", name)
	r, err := http.NewRequest(http.MethodHead, s.harborUrl.String(), nil)
	if err != nil {
		return nil, errors.NewEf(err, "while building http request")
	}
	r2, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, errors.NewEf(err, "while making request to check if project name already exists")
	}

	if r2.StatusCode == http.StatusOK {
		return fn.NewBool(true), nil
	}
	return fn.NewBool(false), nil
}

func (s *server) CreateHarborProject(ctx context.Context, in *ci.HarborProjectIn) (*ci.HarborProjectOut, error) {
	b, err := s.checkIfProjectExists(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if b != nil && *b {
		return &ci.HarborProjectOut{Status: true}, nil
	}

	body := t.M{
		"project_name": in.Name,
		"public":       false,
	}
	bbody, err := json.Marshal(body)
	if err != nil {
		return nil, errors.NewEf(err, "could not unmarshal req body")
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", s.harborUrl.String(), "projects"), bytes.NewBuffer(bbody))
	if err != nil {
		return nil, errors.NewEf(err, "could not build request")
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(s.harborUsername, s.harborPassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, errors.NewEf(err, "while making request")
	}
	if resp.StatusCode == http.StatusCreated {
		return &ci.HarborProjectOut{Status: true}, nil
	}
	return nil, errors.Newf("could not create harbor project as received (statuscode=%d)", resp.StatusCode)
}

func (s *server) DeleteHarborProject(ctx context.Context, in *ci.HarborProjectIn) (*ci.HarborProjectOut, error) {
	b, err := s.checkIfProjectExists(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if b != nil && *b {
		return &ci.HarborProjectOut{Status: true}, nil
	}
	if b != nil && !*b {
		return nil, errors.Newf("harbor project(name=%s) does not exist", in.Name)
	}

	u, err := s.harborUrl.Parse(in.Name)
	if err != nil {
		return nil, errors.NewEf(err, "could not join url path param")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, errors.NewEf(err, "while building http request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.NewEf(err, "while making request")
	}
	if resp.StatusCode == http.StatusOK {
		return &ci.HarborProjectOut{Status: true}, nil
	}
	return nil, errors.Newf("could not delete harbor project as received (statuscode=%d)", resp.StatusCode)
}

func fxCiServer(env *Env, d domain.Domain) ci.CIServer {
	hUrl, err := url.Parse(env.HarborUrl)
	if err != nil || hUrl == nil {
		panic(fmt.Errorf("harbor url (%s) is not a valid url", env.HarborUrl))
	}
	return &server{
		harborUsername: env.HarborUsername,
		harborPassword: env.HarborPassword,
		harborUrl:      *hUrl,
		d:              d,
	}
}
