package api

import (
	"fmt"

	"k8s.io/api/core/v1"
)

// LogConfig is used to represent the log configuration info of one deployment/statefulset
type LogConfig struct {
	// The name of the object to be collected
	Name string `json:"name"`

	// The namespace of the object to be collect
	Namespace string `json:"namespace"`

	// The kind of the object
	Kind string `json:"kind"`

	// The volumeMount of this object(deployment, statefulset) which holds the log
	VolumeMount string `json:"volume_mount"`

	// The labelselector used to list pods of this deploy
	LabelSelector string `json:"-"`

	// The config of the log of this kind
	Config string `json:"config"`
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
	// The name of pod, who generates the log
	PodName string `json:"pod_name"`

	// The namespace of this pod
	Namespace string `json:"namespace"`

	// The volume mount of this pod which generates this log source.
	VolumeMount string `json:"volume_mount"`

	// The raw config file for this log source
	Config string `json:"config"`
}

type LogSourceStatus struct {
	ConfigStatus ConfigStatus `json:"config_status"`

	LogStatus LogStatus `json:"log_status"`
}

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

func NewLogSource(pod *v1.Pod, config *LogConfig) *LogSource {
	return &LogSource{
		Meta: Meta{
			Name: fmt.Sprintf("%s_%s_%s", config.Name, config.VolumeMount, pod.Name),
		},
		Spec: LogSourceSpec{
			Namespace:   pod.ObjectMeta.Namespace,
			PodName:     pod.ObjectMeta.Name,
			VolumeMount: config.VolumeMount,
			Config:      config.Config,
		},
	}
}

func (l *LogSource) GetLogDir() string {
	return fmt.Sprintf("%s/%s_%s", l.getVolumeMountPath(), l.Spec.Namespace, l.Spec.PodName)
}

func (l *LogSource) GetLogMetaDir() string {
	return fmt.Sprintf("%s/%s_%s/.meta", l.getVolumeMountPath(), l.Spec.Namespace, l.Spec.PodName)
}

func (l *LogSource) getVolumeMountPath() string {
	return fmt.Sprintf("/%s", l.Spec.VolumeMount)
}
