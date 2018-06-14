package logmanager

import (
	"strings"
)

type LogSourceAction string

const (
	LogSourceAdd LogSourceAction = "LogSourceAdd"
	LogSourceDel LogSourceAction = "LogSourceDel"
	LogSourceMov LogSourceAction = "LogSourceMov"
	LogSourceNop LogSourceAction = "LogSourceNop"
)

func judgeAction(m *Match) LogSourceAction {
	if m.PodName != "" && m.AgentName == "" && m.ConfPath == "" {
		return LogSourceAdd
	} else if m.PodName == "" && m.AgentName != "" && m.ConfPath != "" {
		return LogSourceDel
	} else if m.PodName != "" && m.AgentName != "" && m.ConfPath != "" && strings.Index(m.ConfPath, m.AgentName) == -1 {
		return LogSourceMov
	} else {
		return LogSourceNop
	}
}

// Create a config file for new logSource to logAgent
func (lm *LogManager) logSourceAddFunc(key string) (bool, error) {
	logSource := lm.LogSources[key]
	logAgentName := lm.Match[key].AgentName

	err := lm.LogAgentManager.AddConfig(&logSource, logAgentName)
	if err != nil {
		return false, err
	}
	// Add log config file to logAgent
	return true, nil
}

func (lm *LogManager) logSourceDelFunc(key string) (bool, error) {
	// Check the log lag
	logSource := lm.LogSources[key]
	logAgentName := lm.Match[key].AgentName

	if logSource.Status.LogStatus.Done {
		// If the log is done collecting, then delete this logSource and config
		err := lm.LogAgentManager.DelConfig(&logSource, logAgentName)
		if err != nil {
			return false, err
		}
		err = lm.removeLogSource(&logSource)
		if err != nil {
			return false, err
		}

		return true, nil

	} else {
		return false, nil
	}
}

func (lm *LogManager) logSourceMovFunc(key string) (bool, error) {
	// Get old agent name from conf path
	logSource := lm.LogSources[key]
	newLogAgentName := lm.Match[key].AgentName
	logConfPath := lm.Match[key].ConfPath
	oldLogAgentName := lm.LogAgentManager.GetAgentNameFromConf(logConfPath)

	// remove old agent config
	err := lm.LogAgentManager.DelConfig(&logSource, oldLogAgentName)
	if err != nil {
		return false, err
	}

	// add new agent conf
	err = lm.LogAgentManager.AddConfig(&logSource, newLogAgentName)
	if err != nil {
		return false, err
	}

	return true, nil
}
