package entities

import (
	"encoding/json"
	"io"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type PortType string

const (
	PortTypeHTTP PortType = "http"
	PortTypeTCP  PortType = "tcp"
	PortTypeUDP  PortType = "udp"
)

type ExposedPort struct {
	Port       int64    `json:"port" bson:"port"`
	TargetPort int64    `json:"target_port" bson:"target_port"`
	Type       PortType `json:"type" bson:"type"`
}

type AttachedResource struct {
	ResourceId repos.ID `json:"resource_id" bson:"resource_id"`
}

type Limit struct {
	Min string `json:"min" bson:"min"`
	Max string `json:"max" bson:"max"`
}

type EnvVar struct {
	Key    string  `json:"name" bson:"name"`
	Type   string  `json:"type,omitempty"`
	Value  *string `json:"value,omitempty"`
	Ref    *string `json:"ref,omitempty"`
	RefKey *string `json:"key,omitempty"`
}

type VolumeMount struct {
	MountPath string `json:"mount_path" bson:"mount_path"`
	Type      string `json:"type" bson:"type"`
	Ref       string `json:"ref" bson:"ref"`
}

type Container struct {
	Name              string             `json:"name" bson:"name"`
	Image             *string            `json:"image" bson:"image"`
	ImagePullSecret   *string            `json:"pull_secret" bson:"pull_secret"`
	EnvVars           []EnvVar           `json:"env_vars" bson:"env_vars"`
	VolumeMounts      []VolumeMount      `json:"volume_mounts" bson:"volume_mounts"`
	AttachedResources []AttachedResource `json:"attached_resources" bson:"attached_resources"`
	ComputePlan       string             `json:"compute_plan" bson:"compute_plan"`
	Quantity          float64            `json:"quantity" bson:"quantity"`
	IsShared          bool               `json:"is_shared" bson:"is_shared"`
}

// type ContainerIn struct {
// 	Name                string
// 	Image               *string
// 	ImagePullSecret     *string
// 	EnvVars             []EnvVar
// 	ComputePlanName     string
// 	ComputePlanQuantity float64
// 	SharingEnabled      bool
// 	AttachedResources   []AttachedResource
// }

type AppStatus string

const (
	AppStateSyncing    = AppStatus("sync-in-progress")
	AppStateRestarting = AppStatus("restarting")
	AppStateFrozen     = AppStatus("frozen")
	AppStateDeleting   = AppStatus("deleting")
	AppStateLive       = AppStatus("live")
	AppStateError      = AppStatus("error")
	AppStateDown       = AppStatus("down")
)

type AutoScale struct {
	MinReplicas     int64 `json:"min_replicas" bson:"min_replicas"`
	MaxReplicas     int64 `json:"max_replicas" bson:"max_replicas"`
	UsagePercentage int64 `json:"usage_percentage" bson:"usage_percentage"`
}

type App struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	crdsv1.App       `json:",inline" bson:",inline"`
}

func (app *App) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), app); err != nil {
		return err
	}

	// if err := validator.Validate(*app); err != nil {
	//  return err
	// }

	return nil
}

func (app App) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(app)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

type App2 struct {
	repos.BaseEntity  `bson:",inline"`
	IsLambda          bool               `json:"is_lambda" bson:"is_lambda"`
	ReadableId        string             `json:"readable_id" bson:"readable_id"`
	ProjectId         repos.ID           `json:"project_id" bson:"project_id"`
	Name              string             `json:"name" bson:"name"`
	Namespace         string             `json:"namespace" bson:"namespace"`
	Frozen            bool               `json:"frozen" bson:"frozen"`
	InterceptDeviceId *repos.ID          `json:"interceptDeviceId" bson:"intercept_device_id"`
	Description       *string            `json:"description" bson:"description"`
	Replicas          int                `json:"replicas" bson:"replicas"`
	AutoScale         *AutoScale         `json:"auto_scale" bson:"auto_scale"`
	ExposedPorts      []ExposedPort      `json:"exposed_ports" bson:"exposed_ports"`
	Containers        []Container        `json:"containers" bson:"containers"`
	Status            AppStatus          `json:"status" bson:"status"`
	Conditions        []metav1.Condition `json:"conditions" bson:"conditions"`
	Metadata          map[string]any     `json:"metadata" bson:"metadata"`
	Overrides         string             `bson:"overrides,omitempty" json:"overrides,omitempty"`
}

var AppIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
