package graph

import (
	"kloudlite.io/apps/infra/internal/app/graph/model"
	fn "kloudlite.io/pkg/functions"
)

func toModelStatus(status any) (*model.Status, error) {
	var rStatus model.Status
	if err := fn.JsonConversion(status, &rStatus); err != nil {
		return nil, err
	}
	return &rStatus, nil
}
