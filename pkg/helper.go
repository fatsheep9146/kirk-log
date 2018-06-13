package logmanager

import (
	"fmt"

	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
)

func LogConfigKeyFunc(cfg *api.LogConfig) string {
	return fmt.Sprintf("%s_%s", cfg.Name, cfg.VolumeMount)
}

func LogConfigConvertFromSliceToMap(cfgs []api.LogConfig) map[string]api.LogConfig {
	cfgmap := make(map[string]api.LogConfig)

	for _, cfg := range cfgs {
		key := LogConfigKeyFunc(&cfg)
		cfgmap[key] = cfg
	}

	return cfgmap
}

func LogSourceKeyFunc(src *api.LogSource) string {
	return fmt.Sprintf("%s_%s", src.Spec.PodName, src.Spec.VolumeMount)
}

func LogSourceConvertFromSliceToMap(srcs []api.LogSource) map[string]api.LogSource {
	srcmap := make(map[string]api.LogSource)

	for _, src := range srcs {
		key := LogSourceKeyFunc(&src)
		srcmap[key] = src
	}

	return srcmap
}

func LogAgentConvertFromSliceToMap(srcs []agent.Agent) map[string]agent.Agent {
	srcmap := make(map[string]agent.Agent)

	for _, src := range srcs {
		key := src.Name
		srcmap[key] = src
	}

	return srcmap
}