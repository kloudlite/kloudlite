package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/slack-go/slack"
	"go.uber.org/fx"
	"kloudlite.io/apps/slack-notifier/internal/env"
)

func fxSlackRoutes() fx.Option {
	return fx.Invoke(
		func(app *fiber.App, slackApi *slack.Client, ev *env.Env) {
			app.Get(
				"/", func(ctx *fiber.Ctx) error {
					channels, _, err := slackApi.GetConversations(&slack.GetConversationsParameters{})
					if err != nil {
						return err
					}
					return ctx.JSON(channels)
				},
			)

			app.Post(
				"/chat", func(ctx *fiber.Ctx) error {
					message, s, err := slackApi.PostMessage(
						ev.SlackChannelID, slack.MsgOptionCompose(
							slack.MsgOptionText("hi sample", true),
							slack.MsgOptionBlocks(
								slack.SectionBlock{
									Type: slack.MBTSection,
									Text: &slack.TextBlockObject{
										Type:     slack.MarkdownType,
										Text:     "## Hi\n**are you watching this**, _really_",
										Emoji:    false,
										Verbatim: false,
									},
									BlockID:   "",
									Fields:    nil,
									Accessory: nil,
								},
							),
						),
					)
					if err != nil {
						return err
					}
					fmt.Println(s)
					return ctx.JSON(message)
				},
			)

		},
	)
}

var Module = fx.Module(
	"app",
	fxSlackRoutes(),
)
