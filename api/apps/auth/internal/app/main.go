package app

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	CookieDomain     string `env:"COOKIE_DOMAIN" required:"true"`
	GithubWebhookUrl string `env:"GITHUB_WEBHOOK_URL" required:"true"`
	GitlabWebhookUrl string `env:"GITLAB_WEBHOOK_URL" required:"true"`

	GithubClientId     string `env:"GITHUB_CLIENT_ID" required:"true"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET" required:"true"`
	GithubCallbackUrl  string `env:"GITHUB_CALLBACK_URL" required:"true"`
	GithubAppId        string `env:"GITHUB_APP_ID" required:"true"`
	GithubAppPKFile    string `env:"GITHUB_APP_PK_FILE" required:"true"`

	GitlabClientId     string `env:"GITLAB_CLIENT_ID" required:"true"`
	GitlabClientSecret string `env:"GITLAB_CLIENT_SECRET" required:"true"`
	GitlabCallbackUrl  string `env:"GITLAB_CALLBACK_URL" required:"true"`

	GoogleClientId     string `env:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET" required:"true"`
	GoogleCallbackUrl  string `env:"GOOGLE_CALLBACK_URL" required:"true"`
}

func (env *Env) GoogleConfig() (clientId string, clientSecret string, callbackUrl string) {
	return env.GoogleClientId, env.GoogleClientSecret, env.GoogleCallbackUrl
}

func (env *Env) GitlabConfig() (clientId string, clientSecret string, callbackUrl string) {
	return env.GitlabClientId, env.GitlabClientSecret, env.GitlabCallbackUrl
}

func (env *Env) GithubConfig() (clientId, clientSecret, callbackUrl, githubAppId, githubAppPKFile string) {
	return env.GithubClientId, env.GithubClientSecret, env.GithubCallbackUrl, env.GithubAppId, env.GithubAppPKFile
}

type CommsClientConnection *grpc.ClientConn

var Module = fx.Module("app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.User]("users", "usr", domain.UserIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("access_tokens", "tkn", domain.AccessTokenIndexes),
	repos.NewFxMongoRepo[*domain.RemoteLogin]("remote_logins", "rml", domain.RemoteTokenIndexes),
	cache.NewFxRepo[*domain.VerifyToken](),
	cache.NewFxRepo[*domain.ResetPasswordToken](),

	fx.Provide(func(conn CommsClientConnection) comms.CommsClient {
		return comms.NewCommsClient((*grpc.ClientConn)(conn))
	}),

	fx.Provide(fxGithub),
	fx.Provide(fxGitlab),
	fx.Provide(fxGoogle),

	fx.Provide(fxRPCServer),
	fx.Invoke(func(server *grpc.Server, authServer auth.AuthServer) {
		auth.RegisterAuthServer(server, authServer)
	}),

	fx.Invoke(func(
		server *fiber.App,
		d domain.Domain,
		env *Env,
		cacheClient cache.Client,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: graph.NewResolver(d)},
		)
		httpServer.SetupGQLServer(
			server,
			schema,
			httpServer.NewSessionMiddleware[*common.AuthSession](
				cacheClient,
				common.CookieName,
				env.CookieDomain,
				common.CacheSessionPrefix,
			),
		)
	}),
	fx.Invoke(func(server *fiber.App, d domain.Domain) {
		server.Post("/cli/new-remote-login", func(ctx *fiber.Ctx) error {
			type RemoteLoginData struct {
				Secret string `json:"secret"`
			}
			var loginData RemoteLoginData
			err := ctx.JSON(&loginData)
			login, err := d.CreateRemoteLogin(ctx.Context(), loginData.Secret)
			if err != nil {
				return err
			}
			ctx.JSON(login)
			return nil
		})
		server.Post("/cli/remote-login/:loginId", func(c *fiber.Ctx) error {
			loginId := c.Params("loginId")
			type RemoteLoginStatusRequest struct {
				Secret string `json:"secret"`
			}
			var request RemoteLoginStatusRequest
			err := c.JSON(&request)
			if err != nil {
				return err
			}
			login, err := d.GetRemoteLogin(c.Context(), repos.ID(loginId), request.Secret)
			if err != nil {
				return err
			}
			c.JSON(map[string]string{
				"status":      string(login.LoginStatus),
				"auth_header": login.AuthHeader,
			})
			return nil
		})
	}),
	domain.Module,
)
