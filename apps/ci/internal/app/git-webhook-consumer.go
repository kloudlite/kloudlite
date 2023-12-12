package app

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"github.com/kloudlite/api/constants"
	"text/template"
	"time"

	"github.com/kloudlite/api/apps/ci/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/finance"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/redpanda"
	text_templates "github.com/kloudlite/api/pkg/text-templates"
	"github.com/kloudlite/api/pkg/types"
	"go.uber.org/fx"
)

var (
	//go:embed templates
	dirTemplates embed.FS
)

type pt struct {
	tt     *template.Template
	logger logging.Logger
}

func (pt *pt) RenderPipelineRun(pipelineRunParams []*domain.TektonVars) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)
	if err := pt.tt.ExecuteTemplate(
		b, "pipeline-run.yml.tpl", map[string]any{
			"pipeline-runs": pipelineRunParams,
		},
	); err != nil {
		pt.logger.Errorf(err, "error parsing template (pipeline-run.yml.tpl)")
		return nil, err
	}
	return b, nil
}

const (
	GithubEventHeader string = "X-Github-Event"
	GitlabEventHeader string = "X-Gitlab-Event"
)

func fxInvokeProcessGitWebhooks() fx.Option {
	return fx.Options(

		fx.Provide(
			func(logger logging.Logger) (domain.PipelineTemplate, error) {
				t := template.New("taskrun")
				t = text_templates.WithFunctions(t)
				if _, err := t.ParseFS(dirTemplates, "templates/pipeline-run.yml.tpl"); err != nil {
					return nil, err
				}
				return &pt{tt: t, logger: logger}, nil
			},
		),

		fx.Invoke(
			func(d domain.Domain, financeClient finance.FinanceClient, consumer redpanda.Consumer, producer redpanda.Producer,
				logr logging.Logger, t domain.PipelineTemplate) {
				consumer.StartConsuming(
					func(msg []byte, _ time.Time, offset int64) error {
						logger := logr.WithName("ci-webhook").WithKV("offset", offset)
						logger.Infof("started processing")
						defer func() {
							logger.Infof("finished processing")
						}()

						var gitHook types.GitHttpHook
						if err := json.Unmarshal(msg, &gitHook); err != nil {
							logger.Errorf(err, "could not unmarshal into *GitWebhookPayload")
							return err
						}

						hook, err := func() (*domain.GitWebhookPayload, error) {
							if gitHook.GitProvider == constants.ProviderGithub {
								return d.ParseGithubHook(gitHook.Headers[GithubEventHeader], gitHook.Body)
							}
							if gitHook.GitProvider == constants.ProviderGitlab {
								return d.ParseGitlabHook(gitHook.Headers[GitlabEventHeader], gitHook.Body)
							}
							return nil, errors.New("unknown git provider")
						}()
						if err != nil {
							if _, ok := err.(*domain.ErrEventNotSupported); ok {
								logger.Infof(err.Error())
								return nil
							}
							logger.Errorf(err, "could not extract gitHook")
							return err
						}

						logger = logger.WithKV("repo", hook.RepoUrl, "provider", hook.GitProvider, "branch", hook.GitBranch)

						tctx, cancelFn := context.WithTimeout(context.TODO(), 3*time.Second)
						defer cancelFn()

						pipelines, err := d.ListPipelinesByGitInfo(tctx, hook.RepoUrl, hook.GitProvider, hook.GitBranch)
						if err != nil {
							return errors.NewEf(err, "listing pipelines by git info")
						}

						if len(pipelines) == 0 {
							logger.Infof("no pipeline is configured for given hook body")
							return nil
						}

						// tkRuns, err := d.GetTektonRunParams(context.TODO(), hook.GitProvider, hook.RepoUrl, hook.GitBranch)
						// if err != nil {
						// 	logger.Errorf(err, "could not get tekton run params")
						// 	return err
						// }

						// if len(tkRuns) == 0 {
						// 	logger.Infof("no pipeline is configured for given hook body")
						// 	return nil
						// }

						accountRuns := map[string][]*domain.TektonVars{}

						for i := range pipelines {
							pRun, err := d.CreateNewPipelineRun(context.TODO(), pipelines[i].Id)
							if err != nil {
								logger.Errorf(err, "creating new pipeline run")
								// return err
							}
							params, err := d.GetPipelineRunParams(context.TODO(), pipelines[i], pRun)
							if err != nil {
								logger.Errorf(err, "getting pipeline run params")
								// return err
							}
							accountRuns[pipelines[i].AccountId] = append(accountRuns[pipelines[i].AccountId], params)
						}

						// for i := range tkRuns {
						// 	tkRuns[i].GitCommitHash = hook.CommitHash
						// 	accountRuns[tkRuns[i].AccountId] = append(accountRuns[tkRuns[i].AccountId], tkRuns[i])
						// }

						for k := range accountRuns {
							cluster, err := financeClient.GetAttachedCluster(context.TODO(), &finance.GetAttachedClusterIn{AccountId: k})
							if err != nil {
								continue
							}

							b, err := t.RenderPipelineRun(accountRuns[k])
							agentMsgBytes, err := json.Marshal(map[string]any{"action": "create", "yamls": b.Bytes()})
							if err != nil {
								return err
							}

							ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
							topicName := cluster.ClusterId + "-incoming"

							pMsg, err := producer.Produce(ctx, topicName, hook.RepoUrl, agentMsgBytes)
							if err != nil {
								cancelFn()
								logger.Errorf(err, "error processing message, could not pipeline output into topic=%s", topicName)
								return err
							}
							logger.Infof("processed git webhook, pipelined output into to topic=%s, offset=%d", pMsg.Topic, pMsg.Offset)
							cancelFn()
						}
						return nil
					},
				)
			},
		),
	)
}
