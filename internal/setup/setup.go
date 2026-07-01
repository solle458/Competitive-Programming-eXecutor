package setup

import "Competitive-Programming-eXecutor/internal/config"

type Request struct {
	ContestID  string
	Lang       string
	WorkingDir string
	Config     *config.Config
}

type Provider interface {
	Supports(contestID string) bool
	Setup(req Request) error
}
