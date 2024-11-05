package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
import (
	"github.com/kloudlite/api/apps/message-office/internal/domain"
)

type Resolver struct {
	Domain domain.Domain
}
