package dto

import (
	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/pkg/apis/platform/v1alpha1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ListResponse represents a standard list response
type ListResponse struct {
	Items interface{} `json:"items"`
	Count int         `json:"count"`
}

// Environment Responses

type EnvironmentResponse struct {
	Message     string                      `json:"message"`
	Environment *environmentsv1.Environment `json:"environment"`
}

type EnvironmentListResponse struct {
	Environments []environmentsv1.Environment `json:"environments"`
	Count        int                          `json:"count"`
}

type EnvironmentStatusResponse struct {
	Name            string                           `json:"name"`
	Namespace       string                           `json:"namespace"`
	Activated       bool                             `json:"activated"`
	Status          environmentsv1.EnvironmentStatus `json:"status"`
	ResourceQuotas  *environmentsv1.ResourceQuotas   `json:"resourceQuotas,omitempty"`
	NetworkPolicies *environmentsv1.NetworkPolicies  `json:"networkPolicies,omitempty"`
}

// Composition Responses

type CompositionResponse struct {
	Message     string                      `json:"message"`
	Composition *environmentsv1.Composition `json:"composition"`
}

type CompositionListResponse struct {
	Compositions []environmentsv1.Composition `json:"compositions"`
	Count        int                          `json:"count"`
}

type CompositionStatusResponse struct {
	Name             string                          `json:"name"`
	Namespace        string                          `json:"namespace"`
	State            environmentsv1.CompositionState `json:"state"`
	Message          string                          `json:"message"`
	ServicesCount    int32                           `json:"servicesCount"`
	RunningCount     int32                           `json:"runningCount"`
	Services         []environmentsv1.ServiceStatus  `json:"services,omitempty"`
	Endpoints        map[string]string               `json:"endpoints,omitempty"`
	LastDeployedTime *metav1.Time                    `json:"lastDeployedTime,omitempty"`
}

// Service Responses

type ServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
}

type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"clusterIP"`
	Ports     []ServicePort     `json:"ports"`
	Selector  map[string]string `json:"selector,omitempty"`
}

type ServiceListResponse struct {
	Services []ServiceInfo `json:"services"`
	Count    int           `json:"count"`
}

// User Responses

type UserResponse struct {
	Message string                 `json:"message"`
	User    *platformv1alpha1.User `json:"user"`
}

type UserListResponse struct {
	Users []platformv1alpha1.User `json:"users"`
	Count int                     `json:"count"`
}

// MachineType Responses

type MachineTypeResponse struct {
	Message string                  `json:"message,omitempty"`
	Success bool                    `json:"success,omitempty"`
	Data    *machinesv1.MachineType `json:"data,omitempty"`
	Active  bool                    `json:"active,omitempty"`
}

type MachineTypeListResponse struct {
	MachineTypes []machinesv1.MachineType `json:"machineTypes"`
	Count        int                      `json:"count"`
}

// WorkMachine Responses

type WorkMachineResponse struct {
	Message string                  `json:"message,omitempty"`
	Success bool                    `json:"success,omitempty"`
	Data    *machinesv1.WorkMachine `json:"data,omitempty"`
}

type WorkMachineListResponse struct {
	WorkMachines []machinesv1.WorkMachine `json:"workMachines"`
	Count        int                      `json:"count"`
}

// Workspace Responses

type WorkspaceResponse struct {
	Message   string                  `json:"message"`
	Workspace *workspacesv1.Workspace `json:"workspace"`
}

type WorkspaceListResponse struct {
	Workspaces []workspacesv1.Workspace `json:"workspaces"`
	Count      int                      `json:"count"`
}

// Health Responses

type HealthResponse struct {
	Status string `json:"status"`
	Time   int64  `json:"time"`
}

// Auth Responses

type ValidateTokenResponse struct {
	Valid bool        `json:"valid"`
	User  interface{} `json:"user"`
}

// Info Response

type InfoResponse struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
}
