package app

import (
	"github.com/fatsheep9146/kirklog/pkg"
)

type LogManagerServer struct {
	Cfg *logmanager.LogManagerConfig
}

func NewLogManagerServer() *LogManagerServer {
	cfg := logmanager.NewLogManagerConfig()

	return &LogManagerServer{
		Cfg: cfg,
	}
}

func (s *LogManagerServer) Run() error {
	done := make(chan struct{})

	// New and start the logManager
	go logmanager.NewLogManager(s.Cfg).Run()

	<-done
	return nil
}
