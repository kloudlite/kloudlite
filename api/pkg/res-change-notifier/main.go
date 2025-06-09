package rcn

import (
	"fmt"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"net/http"
)

type ResourceChangeNotifier interface {
	Notify(resourceId repos.ID) error
}

type rcn struct {
	serviceUrl string
}

func (r *rcn) Notify(resourceId repos.ID) error {
	_, err := http.Post(fmt.Sprintf("%s/publish/resource_update/%s", r.serviceUrl, resourceId), "text/plain", nil)
	fmt.Println(err)
	return errors.NewE(err)
}

func NewResourceChangeNotifier(serviceUrl string) ResourceChangeNotifier {
	return &rcn{
		serviceUrl: serviceUrl,
	}
}

type ResourceChangeNotifierConfig interface {
	GetNotifierUrl() string
}

func NewFxResourceChangeNotifier[T ResourceChangeNotifierConfig]() fx.Option {
	return fx.Module("rcn",
		fx.Provide(func(c T) ResourceChangeNotifier {
			return NewResourceChangeNotifier(c.GetNotifierUrl())
		}),
	)
}
