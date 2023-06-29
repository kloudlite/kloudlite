package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.uber.org/fx"

	"kloudlite.io/apps/auth/internal/env"
	"kloudlite.io/apps/auth/internal/framework"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
)

// nnoremap <C-x> :lua require('fzf-lua').lsp_live_workspace_symbols({ fzf_opts = {['--delimiter'] = ':', ['--with-nth'] = "4,1" } })<CR>
// lua require('fzf-lua').lsp_live_workspace_symbols({ fzf_opts = {['--layout'] = 'reverse-list'} })

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()
	app := fx.New(
		fx.NopLogger,
		fn.FxErrorHandler(),
		fx.Provide(func() (*env.Env, error) {
			return env.LoadEnv()
		}),
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "auth", Dev: isDev})
			},
		),
		framework.Module,
	)

	ctx, cancelFunc := func() (context.Context, context.CancelFunc) {
		if isDev {
			return context.WithTimeout(context.TODO(), 20*time.Second)
		}
		return context.WithTimeout(context.TODO(), 5*time.Second)
	}()
	defer cancelFunc()

	if err := app.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)

	<-app.Done()
}
