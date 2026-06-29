package app

import "Competitive-Programming-eXecutor/internal/config"

type App struct {
	Config *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		Config: cfg,
	}
}
