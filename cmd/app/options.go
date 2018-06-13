package app

import (
	"github.com/spf13/pflag"
)

func (s *LogManagerServer) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Cfg.LogConfigDir, "log-config-dir", "", "The dir where to store the log config files")
}
