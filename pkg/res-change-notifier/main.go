package rcn

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/pkg/repos"
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
	return err
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
