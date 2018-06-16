package app

import (
	"github.com/spf13/pflag"
)

func (s *LogManagerServer) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Cfg.LogConfigDir, "log-config-dir", "", "The dir where to store the log config files")
	fs.StringVar(&s.Cfg.Name, "name", "", "The name of logmanager instance")
	fs.StringVar(&s.Cfg.Namespace, "namespace", "", "The namespace of logmanger instance")
	fs.StringVar(&s.Cfg.AgentType, "agent-type", "logkit", "the agent type that used to collect logs")
	fs.IntVar(&s.logLevel, "log-level", 5, "the log level, [0]:Panic, [1]:Fatal, [2]:Error, [3]:Error, [4]:Warn, [5]:Info, [6]:Debug, default is info")
}
