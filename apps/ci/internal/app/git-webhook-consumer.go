package app

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"text/template"
	"time"

	"kloudlite.io/apps/ci/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	text_templates "kloudlite.io/pkg/text-templates"
	"kloudlite.io/pkg/types"
)

var (
	//go:embed templates
	res embed.FS
)

const (
	GithubEventHeader string = "X-Github-Event"
	GitlabEventHeader string = "X-Gitlab-Event"
)

func ProcessWebhooks(d domain.Domain, consumer redpanda.Consumer, producer redpanda.Producer, logr logging.Logger, env *Env) error {
	t := template.New("taskrun")
	t = text_templates.WithFunctions(t)
	// if _, err := t.ParseFS(res, "templates/taskrun.tpl.yml"); err != nil {
	if _, err := t.ParseFS(res, "templates/pipeline-run.yml.tpl"); err != nil {
		return err
	}

	consumer.StartConsuming(
		func(msg []byte, timeStamp time.Time, offset int64) error {
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
				if gitHook.GitProvider == common.ProviderGithub {
					return d.ParseGithubHook(gitHook.Headers[GithubEventHeader], gitHook.Body)
				}
				if gitHook.GitProvider == common.ProviderGitlab {
					return d.ParseGitlabHook(gitHook.Headers[GitlabEventHeader], gitHook.Body)
				}
				return nil, errors.New("unknown git provider")
			}()
			if err != nil {
				logger.Errorf(err, "could not extract gitHook")
				return err
			}

			logger = logger.WithKV(
				"repo", hook.RepoUrl,
				"provider", hook.GitProvider,
				"branch", hook.GitBranch,
			)

			tkRuns, err := d.GetTektonRunParams(context.TODO(), hook.GitProvider, hook.RepoUrl, hook.GitBranch)
			if err != nil {
				logger.Errorf(err, "could not get tekton run params")
				return err
			}

			if len(tkRuns) == 0 {
				logger.Infof("no pipeline is configured for given hook body")
				return nil
			}

			for i := range tkRuns {
				tkRuns[i].GitCommitHash = hook.CommitHash
			}

			b := new(bytes.Buffer)
			if err := t.ExecuteTemplate(
				// b, "taskrun.tpl.yml", map[string]any{"tekton-runs": tkRuns},
				b, "pipeline-run.yml.tpl", map[string]any{
					"tekton-runs": tkRuns,
				},
			); err != nil {
				logger.Errorf(err, "error parsing template (taskrun.tpl.yml)")
				return err
			}

			agentMsgBytes, err := json.Marshal(map[string]any{"action": "create", "yamls": b.Bytes()})
			if err != nil {
				return err
			}

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			pMsg, err := producer.Produce(ctx, env.KafkaApplyYamlTopic, hook.RepoUrl, agentMsgBytes)
			if err != nil {
				logger.Errorf(err, "error processing message, could not pipeline output into topic=%s", env.KafkaApplyYamlTopic)
				return err
			}
			logger.Infof("processed git webhook, pipelined output into to topic=%s, offset=%d", pMsg.Topic, pMsg.Offset)
			return nil
		},
	)
	return nil
}
