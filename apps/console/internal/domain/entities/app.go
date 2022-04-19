package entities

import "kloudlite.io/pkg/repos"

type PortType string

const (
	PortTypeHTTP PortType = "http"
	PortTypeTCP  PortType = "tcp"
	PortTypeUDP  PortType = "udp"
)

type ExposedPort struct {
	Port       int      `json:"port" bson:"port"`
	TargetPort int      `json:"target_port" bson:"target_port"`
	Type       PortType `json:"type" bson:"type"`
}

type AttachedResource struct {
	ResourceId repos.ID          `json:"resource_id" bson:"resource_id"`
	EnvVars    map[string]string `json:"env_vars" bson:"env_vars"`
}

type Limit struct {
	Min string `json:"min" bson:"min"`
	Max string `json:"max" bson:"max"`
}

type Container struct {
	Name         string            `json:"name" bson:"name"`
	Image        string            `json:"image" bson:"image"`
	Command      []string          `json:"command" bson:"command"`
	Args         []string          `json:"args" bson:"args"`
	EnvVars      map[string]string `json:"env_vars" bson:"env_cars"`
	CPULimits    Limit             `json:"cpu_limits" bson:"cpu_limits"`
	MemoryLimits Limit             `json:"memory_limits" bson:"memory_limits"`
}

type AppStatus string

const (
	AppStateSyncing = AppStatus("sync-in-progress")
	AppStateLive    = AppStatus("live")
	AppStateError   = AppStatus("error")
	AppStateDown    = AppStatus("down")
)

type App struct {
	repos.BaseEntity  `bson:",inline"`
	Name              string             `json:"name" bson:"name"`
	Namespace         string             `json:"namespace" bson:"namespace"`
	Description       string             `json:"description" bson:"description"`
	Replicas          int                `json:"replicas" bson:"replicas"`
	ExposedPorts      []ExposedPort      `json:"exposed_ports" bson:"exposed_ports"`
	AttachedResources []AttachedResource `json:"attached_resources" bson:"attached_resources"`
	Containers        []Container        `json:"containers" bson:"containers"`
	Status            ClusterStatus      `json:"status" bson:"status"`
}
