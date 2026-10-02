package main

import (
	"Competitive-Programming-eXecutor/cmd"
	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/config"
)

func main() {
	cfg := config.NewConfig()
	app := app.NewApp(cfg)
	cmd.Execute(app)
}
