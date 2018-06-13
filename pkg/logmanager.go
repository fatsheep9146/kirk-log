package logmanager

import (
	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
	"github.com/fatsheep9146/kirklog/pkg/logkit"
)

type LogManagerConfig struct {
	LogConfigDir string `json:"log_config_dir"`
}

type LogManager struct {
	// map for all log config
	LogConfigs map[string]api.LogConfig

	// map for all log entries
	LogSources map[string]api.LogSource

	// map for all log agent
	LogAgents map[string]agent.Agent

	// the working queue to store the logSource wait to be processed
	Queue string

	// the Agent used to manage the log agent components
	LogAgentManager agent.AgentManager
}

func NewLogManagerConfig() *LogManagerConfig {
	return &LogManagerConfig{}
}

func NewLogManager(cfg *LogManagerConfig) *LogManager {
	// Register serveral functions for different types of log agent

	// Load logConfigs from files
	logConfigs := LoadLogConfig(cfg.LogConfigDir)
	// if err != nil {
	// 	panic(err)
	// }

	// Create logsources spec map from logConfigs
	logSources := CreateLogSourcesFromLogConfigs(logConfigs)

	// Check and create log agent configs
	logAgentManager := NewAgentManager(logConfigs[0].Agent)
	logAgents := logAgentManager.List()
	if len(logAgents) == 0 {
		logAgentManager.Deploy()
	}

	// Restore the logsources map status from current situations in case this is a restart

	return &LogManager{
		LogConfigs:      LogConfigConvertFromSliceToMap(logConfigs),
		LogSources:      LogSourceConvertFromSliceToMap(logSources),
		LogAgents:       LogAgentConvertFromSliceToMap(logAgents),
		LogAgentManager: logAgentManager,
	}
}

// 1. Start the loop function which can find the new added pods, delete pods, new added agent pods, new delete agent pods
// 2. Start the workers routine
// 3. Start the log collect checker routine to check the collection lag of each runner
func (lm *LogManager) Run() {
	//

}

// The worker routine to sync the
func (lm *LogManager) sync(key string) {

}

// Create logconfig objects from the files under the path dir
func LoadLogConfig(path string) []api.LogConfig {
	logConfigs := make([]api.LogConfig, 0)

	return logConfigs
}

func CreateLogSourcesFromLogConfigs([]api.LogConfig) []api.LogSource {
	logsources := make([]api.LogSource, 0)

	return logsources
}

func NewAgentManager(t agent.AgentType) agent.AgentManager {
	switch t {
	case agent.Logkit:
		return logkit.NewLogkitAgent()
	default:
		return nil
	}
}
