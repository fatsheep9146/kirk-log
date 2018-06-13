package api

import (
	"github.com/fatsheep9146/kirklog/pkg/agent"
)

// LogConfig is used to represent the log configuration info of one deployment/statefulset
type LogConfig struct {
	// The name of the object to be collected
	Name string `json:"name"`

	// The kind of the object
	Kind string `json:"kind"`

	// The volumeMount of this object(deployment, statefulset) which holds the log
	VolumeMount string `json:"volume_mount"`

	// The config of the log of this kind
	Config string `json:"config"`

	// The type of log collector agent
	Agent agent.AgentType `json:"agent"`
}

// This object is used to repesent one log config from one pod of one deployment/statefulset
// For example, the app log of the pod boots-gate-xxx which belongs to the deployment "boots-gate"

type LogSource struct {
	Meta   Meta            `json:"meta"`
	Spec   LogSourceSpec   `json:"spec"`
	Status LogSourceStatus `json:"status"`
}

type Meta struct {
	Name string `json:"name"`
}

type LogSourceSpec struct {
	// The namespace of this pod
	Namespace string `json:"namespace"`

	// The name of pod, who generates the log
	PodName string `json:"pod_name"`

	// Pod Status
	PodStatus PodStatus `json:"pod_status"`

	// The volume mount of this pod which generates this log source.
	VolumeMount string `json:"volume_mount"`

	// The name of the log agent who is responsable for collecting this log source
	LogAgent string `json:"log_agent"`

	// The raw config file for this log source
	Config string `json:"config"`
}

type LogSourceStatus struct {
	ConfigStatus ConfigStatus `json:"config_status"`

	LogStatus LogStatus `json:"log_status"`
}

type PodStatus string

// The status of the config file path
type ConfigStatus struct {
	// The path of config file for this log source
	Path string `json:"path"`
}

// The status of the log file collection
type LogStatus struct {
	// The flag indicates that the log is done collecting
	Done bool `json:"done"`
}
