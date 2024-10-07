/*
 * Copyright (C) 2024 by Jason Figge
 */

package application

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/resources/engine"
	engineModels "us.figge.auto-ssh/internal/resources/models"
)

type Application struct {
	stats   engineModels.Stats
	hosts   engineModels.HostEngine
	tunnels engineModels.TunnelEngine
}

func NewApplication(
	hosts engineModels.HostEngine,
	tunnels engineModels.TunnelEngine,
) *Application {
	a := &Application{
		hosts:   hosts,
		tunnels: tunnels,
		stats:   engine.NewStatsEngine(),
	}

	//log.InitLogManager(
	//	log.LogOptionSize(config.C.Log.Size),
	//	log.LogOptionTTL(a.cm.config.Log.TTL),
	//)
	return a
}

func (a *Application) Start(ctx context.Context) sync {
	a.stats.StartStatsTunnel(ctx, config.C.Monitor.StatsPort)
	wg := a.tunnels.StartTunnels(ctx)
	return n
}
