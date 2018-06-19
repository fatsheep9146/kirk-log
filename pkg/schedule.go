package logmanager

import (
	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
	log "github.com/sirupsen/logrus"
)

// Currently we schedule the logSource to the agent with the smallest count of logSources
func schedule(logsource *api.LogSource, logAgentsMap map[string]*agent.Agent, match map[string]*Match) {
	logger := log.WithFields(log.Fields{
		"func": "schedule",
	})
	logger.Infof("Start scheduling the logsource %s", logsource.Meta.Name)
	// logger.Debug("Candidate log agent includes:")
	// for _, agent := range logAgentsMap {
	// 	logger.Debugf("  Candidate log agent %s", agent.Name)
	// }

	counts := make(map[string]int)

	for k, a := range logAgentsMap {
		count := 0
		for _, m := range match {
			if m.AgentName == a.Name {
				count++
			}
		}
		counts[k] = count
	}
	// for k, count := range counts {
	// 	logger.Debugf("The count of agent %s is %d", k, count)
	// }

	var minCount int
	var minAgent string = ""

	for k, v := range counts {
		if minAgent == "" {
			minAgent = k
			minCount = v
		} else {
			if minCount > counts[k] {
				minAgent = k
				minCount = v
			}
		}
	}

	logger.Infof("Log agent %s is chosen", minAgent)

	match[logsource.Meta.Name].AgentName = minAgent
}
