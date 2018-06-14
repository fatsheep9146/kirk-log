package logkit

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/fatsheep9146/kirklog/pkg/agent"
)

type LogkitAgentManagerImpl struct {
	Cli  *kubernetes.Clientset
	Name string
}

func NewLogkitAgentManager(cfg *agent.AgentManagerConfig) agent.AgentManager {

	fmt.Printf("Not Implemented yet")

	return nil
}

func (l *LogkitAgentManagerImpl) List() []agent.Agent {
	agents := make([]agent.Agent, 0)

	fmt.Printf("Not Implemented yet")

	return agents
}

func (l *LogkitAgentManagerImpl) Deploy() error {

	fmt.Printf("Not Implemented yet")

	return nil
}
