package logmanager

import (
	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
)

// Currently we schedule the logSource to the agent with the smallest count of logSources
func schedule(logsource api.LogSource, logAgentsMap map[string]agent.Agent, match map[string]*Match) {
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

	match[logsource.Meta.Name].AgentName = minAgent
}
