package agent

type AgentType string

const (
	Logkit      AgentType = "logkit"
	LogExporter AgentType = "logexporter"
	PiliDsync   AgentType = "pili-dsync"
)

type AgentManager interface {
	// This function is use to deploy a log agent deployment
	Deploy() error

	// List the current agent pods of this deployment
	List() []Agent
}

type Agent struct {

	// The pod name of this log agent instance
	Name string `json:"name"`

	// The config path for this agent to read log source config
	Path string `json:"path"`
}
