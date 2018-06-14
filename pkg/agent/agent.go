package agent

import (
	"k8s.io/client-go/kubernetes"

	"github.com/fatsheep9146/kirklog/pkg/api"
)

type AgentType string

const (
	Logkit      AgentType = "logkit"
	LogExporter AgentType = "logexporter"
	PiliDsync   AgentType = "pili-dsync"
)

type AgentManagerConfig struct {
	// The name of LogAgentManager, which is also used to create the log agent components.
	Name string

	// The logConfigs that the logManager needed to collect
	LogConfigs []api.LogConfig

	Cli *kubernetes.Clientset
}

type AgentManager interface {
	// This function is use to deploy a log agent deployment
	Deploy() error

	// List the pods info of log agents
	List() ([]Agent, error)

	// Add the log config of one logSource to one logAgent
	AddConfig(logSource *api.LogSource, agent string) error

	// Delete the log config of one logSource from one logAgent
	DelConfig(logSource *api.LogSource, agent string) error

	// Check the log collect of one logSource from one logAgent
	CheckLag(logSource *api.LogSource, agent string) bool

	// Get agent name from config path
	GetAgentNameFromConf(confpath string) string
}

type Agent struct {

	// The pod name of this log agent instance
	Name string `json:"name"`

	// The config path for this agent to read log source config
	Path string `json:"path"`
}
