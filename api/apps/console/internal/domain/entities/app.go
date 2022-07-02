package entities

import (
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

type Container struct {
	PipelineId        repos.ID           `json:"pipeline_id" bson:"pipeline_id"`
	Name              string             `json:"name" bson:"name"`
	Image             *string            `json:"image" bson:"image"`
	ImagePullSecret   *string            `json:"pull_secret" bson:"pull_secret"`
	EnvVars           []EnvVar           `json:"env_vars" bson:"env_cars"`
	CPULimits         Limit              `json:"cpu_limits" bson:"cpu_limits"`
	MemoryLimits      Limit              `json:"memory_limits" bson:"memory_limits"`
	AttachedResources []AttachedResource `json:"attached_resources" bson:"attached_resources"`
}

type ContainerIn struct {
	Name                string
	Image               *string
	ImagePullSecret     *string
	EnvVars             []EnvVar
	ComputePlanName     string
	ComputePlanQuantity float64
	SharingEnabled      bool
	AttachedResources   []AttachedResource
}

type AppStatus string

const (
	AppStateSyncing = AppStatus("sync-in-progress")
	AppStateLive    = AppStatus("live")
	AppStateError   = AppStatus("error")
	AppStateDown    = AppStatus("down")
)

type AppIn struct {
	ReadableId   string
	ProjectId    repos.ID
	Name         string
	Namespace    string
	Description  *string
	Replicas     int
	ExposedPorts []ExposedPort
	Containers   []ContainerIn
	Status       AppStatus
	Conditions   []metav1.Condition
	Provider     string
	Region       string
}

type App struct {
	repos.BaseEntity `bson:",inline"`
	ReadableId       string             `json:"readable_id" bson:"readable_id"`
	ProjectId        repos.ID           `json:"project_id" bson:"project_id"`
	Name             string             `json:"name" bson:"name"`
	Namespace        string             `json:"namespace" bson:"namespace"`
	Description      *string            `json:"description" bson:"description"`
	Replicas         int                `json:"replicas" bson:"replicas"`
	ExposedPorts     []ExposedPort      `json:"exposed_ports" bson:"exposed_ports"`
	Containers       []Container        `json:"containers" bson:"containers"`
	Status           AppStatus          `json:"status" bson:"status"`
	Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
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
			{Key: "name", Value: repos.IndexAsc},
			{Key: "namespace", Value: repos.IndexAsc},
			{Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
