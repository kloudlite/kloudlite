package tekton_interceptor

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/tekton-interceptor/internal/framework"
)

func main() {
	fx.New(framework.Module).Run()
}
