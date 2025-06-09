package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	fn "github.com/kloudlite/api/pkg/functions"
	t "github.com/kloudlite/api/pkg/types"
	json_patch "github.com/kloudlite/operator/pkg/json-patch"
	"github.com/kloudlite/operator/pkg/operator"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	MetadataResolver   struct{}
	StatusResolver     struct{}
	MetadataInResolver struct{}
)

type PatchResolver struct{}

func (r *PatchResolver) Value(ctx context.Context, obj *json_patch.PatchOperation) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.Value.MarshalJSON()
}

type PatchInResolver struct{}

func (r *PatchInResolver) Value(ctx context.Context, obj *json_patch.PatchOperation, data interface{}) error {
	if obj == nil {
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &obj.Value)
}

func (*MetadataInResolver) Labels(ctx context.Context, obj *v1.ObjectMeta, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Labels)
}

func (*MetadataInResolver) Annotations(ctx context.Context, obj *v1.ObjectMeta, data map[string]interface{}) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Annotations)
}

type SyncStatusResolver struct{}

func (*SyncStatusResolver) Action(ctx context.Context, obj *t.SyncStatus) (t.SyncAction, error) {
	if obj == nil {
		return t.SyncAction(""), fmt.Errorf("syncStatus can not be nil")
	}
	return obj.Action, nil
}

func (*SyncStatusResolver) State(ctx context.Context, obj *t.SyncStatus) (t.SyncState, error) {
	if obj == nil {
		return t.SyncState(""), fmt.Errorf("syncStatus can not be nil")
	}
	return obj.State, nil
}

func (*SyncStatusResolver) SyncScheduledAt(ctx context.Context, obj *t.SyncStatus) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("syncStatus can not be nil")
	}
	return obj.SyncScheduledAt.Format(time.RFC3339), nil
}

func (*SyncStatusResolver) LastSyncedAt(ctx context.Context, obj *t.SyncStatus) (*string, error) {
	if obj == nil {
		return nil, fmt.Errorf("syncStatus can not be nil")
	}
	return fn.New(obj.LastSyncedAt.Format(time.RFC3339)), nil
}

func (*StatusResolver) Checks(ctx context.Context, obj *operator.Status) (map[string]interface{}, error) {
	if obj == nil || obj.Checks == nil {
		return nil, nil
	}
	m := make(map[string]any, len(obj.Checks))
	if err := fn.JsonConversion(obj.Checks, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (*StatusResolver) DisplayVars(ctx context.Context, obj *operator.Status) (map[string]interface{}, error) {
	return nil, nil
}

// 	if obj == nil || obj.DisplayVars == nil {
// 		return nil, nil
// 	}
// 	var m map[string]any
// 	b, err := obj.DisplayVars.MarshalJSON()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := json.Unmarshal(b, &m); err != nil {
// 		return nil, err
// 	}
// 	return m, nil
// }

func (*MetadataResolver) Labels(ctx context.Context, obj *v1.ObjectMeta) (map[string]interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	var m map[string]any
	if err := fn.JsonConversion(obj.Labels, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (*MetadataResolver) Annotations(ctx context.Context, obj *v1.ObjectMeta) (map[string]interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	var m map[string]any
	if err := fn.JsonConversion(obj.Annotations, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (*MetadataResolver) CreationTimestamp(ctx context.Context, obj *v1.ObjectMeta) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("object can not be nil")
	}
	return obj.CreationTimestamp.Format(time.RFC3339), nil
}

func (*MetadataResolver) DeletionTimestamp(ctx context.Context, obj *v1.ObjectMeta) (*string, error) {
	d := obj.GetDeletionTimestamp()
	if d == nil {
		return nil, nil
	}
	return fn.New(d.Format(time.RFC3339)), nil
}
