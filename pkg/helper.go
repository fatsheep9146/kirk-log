package logmanager

import (
	"fmt"

	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
)

func logConfigKeyFunc(cfg *api.LogConfig) string {
	return fmt.Sprintf("%s_%s_%s", cfg.Kind, cfg.Name, cfg.VolumeMount)
}

func logConfigConvertFromSliceToMap(cfgs []api.LogConfig) map[string]*api.LogConfig {
	cfgmap := make(map[string]*api.LogConfig)

	for i, cfg := range cfgs {
		key := logConfigKeyFunc(&cfg)
		cfgmap[key] = &cfgs[i]
	}

	return cfgmap
}

func logSourceConvertFromSliceToMap(srcs []api.LogSource) map[string]*api.LogSource {
	srcmap := make(map[string]*api.LogSource)

	for i, src := range srcs {
		srcmap[src.Meta.Name] = &srcs[i]
	}

	return srcmap
}

func logAgentConvertFromSliceToMap(srcs []agent.Agent) map[string]*agent.Agent {
	srcmap := make(map[string]*agent.Agent)

	for i, src := range srcs {
		key := src.Name
		srcmap[key] = &srcs[i]
	}

	return srcmap
}
