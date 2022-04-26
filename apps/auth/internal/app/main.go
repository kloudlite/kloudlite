package app

import (
	"context"
	"google.golang.org/grpc"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"net/http"

	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/common"
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

type authGrpcServerImpl struct {
	auth.UnimplementedAuthServer
	d domain.Domain
}

func (a *authGrpcServerImpl) GetAccessToken(ctx context.Context, request *auth.GetAccessTokenRequest) (*auth.AccessTokenOut, error) {
	token, err := a.d.GetAccessToken(ctx, request.Provider, request.UserId)
	if err != nil {
		return nil, err
	}
	return &auth.AccessTokenOut{
		UserId:   string(token.UserId),
		Email:    token.Email,
		Provider: token.Provider,
		OauthToken: &auth.OauthToken{
			AccessToken:  token.Token.AccessToken,
			RefreshToken: token.Token.RefreshToken,
			Expiry:       token.Token.Expiry.UnixMilli(),
		},
	}, err
}

func fxRPCServer(d domain.Domain) auth.AuthServer {
	return &authGrpcServerImpl{
		d: d,
	}
}

var Module = fx.Module("app",
	config.EnvFx[Env](),
	repos.NewFxMongoRepo[*domain.User]("users", "usr", domain.UserIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("access_tokens", "tkn", domain.AccessTokenIndexes),
	cache.NewFxRepo[*domain.VerifyToken](),
	cache.NewFxRepo[*domain.ResetPasswordToken](),

	fx.Provide(fxGithub),
	fx.Provide(fxGitlab),
	fx.Provide(fxGoogle),

	fx.Provide(fxRPCServer),
	fx.Invoke(func(server *grpc.Server, authServer auth.AuthServer) {
		auth.RegisterAuthServer(server, authServer)
	}),

	fx.Provide(fxMessenger),
	fx.Invoke(func(
		server *http.ServeMux,
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
			cache.NewSessionRepo[*common.AuthSession](
				cacheClient,
				common.CookieName,
				env.CookieDomain,
				common.CacheSessionPrefix,
			),
		)
	}),
	domain.Module,
)
