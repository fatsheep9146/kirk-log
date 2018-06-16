package app

import (
	log "github.com/sirupsen/logrus"

	"github.com/fatsheep9146/kirklog/pkg"
)

type LogManagerServer struct {
	Cfg      *logmanager.LogManagerConfig
	logLevel int
}

func NewLogManagerServer() *LogManagerServer {
	cfg := logmanager.NewLogManagerConfig()

	return &LogManagerServer{
		Cfg: cfg,
	}
}

func (s *LogManagerServer) Run() error {
	done := make(chan struct{})

	// Initialize the log config
	initLog(s.logLevel)

	// New and start the logManager
	go logmanager.NewLogManager(s.Cfg).Run()

	<-done
	return nil
}

func initLog(level int) {
	switch level {
	case 0:
		log.SetLevel(log.PanicLevel)
	case 1:
		log.SetLevel(log.FatalLevel)
	case 2:
		log.SetLevel(log.ErrorLevel)
	case 3:
		log.SetLevel(log.WarnLevel)
	case 4:
		log.SetLevel(log.InfoLevel)
	case 5:
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}
