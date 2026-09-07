package app

import (
	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/server"
)

type Container struct {
	Config          config.Config
	Server          *server.Server
	FirestoreStatus string
	Close           func()
}

func Build(cfg config.Config) (*Container, error) {
	repositories, err := BuildRepositories(cfg)
	if err != nil {
		return nil, err
	}
	modules, err := BuildModules(cfg, repositories)
	if err != nil {
		repositories.Close()
		return nil, err
	}
	handlers := BuildHandlers(cfg, repositories, modules)
	routes := handlers.Routes(modules.TokenService, cfg.Env, repositories.FirestoreStatus)

	return &Container{
		Config:          cfg,
		Server:          server.New(cfg, routes),
		FirestoreStatus: repositories.FirestoreStatus,
		Close:           repositories.Close,
	}, nil
}
