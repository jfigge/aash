/*
 * Copyright (C) 2024 by Jason Figge
 */

package rest

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	managerModels "us.figge.auto-ssh/internal/web/models"
)

type TunnelRest struct {
	manager managerModels.Tunnel
	routes  []*mux.Router
}

func NewTunnelRest(ctx context.Context, manager managerModels.Tunnel, router *mux.Router) {
	apis := &TunnelRest{
		manager: manager,
	}
	rootRoute := router.PathPrefix("/tunnels")
	rootRoute.Methods(http.MethodGet).HandlerFunc(apis.ListTunnels)
	rootRoute.Methods(http.MethodPost).HandlerFunc(apis.AddTunnel)
	tunnelRoute := rootRoute.PathPrefix("/{tunnel}")
	tunnelRoute.Methods(http.MethodGet).HandlerFunc(apis.GetTunnel)
	tunnelRoute.Methods(http.MethodPut).HandlerFunc(apis.UpdateTunnel)
	tunnelRoute.Methods(http.MethodDelete).HandlerFunc(apis.RemoveTunnel)
	startRoute := tunnelRoute.PathPrefix("/start")
	startRoute.Methods(http.MethodPut).HandlerFunc(apis.StartTunnel)
	stopRoute := tunnelRoute.PathPrefix("/start")
	stopRoute.Methods(http.MethodPut).HandlerFunc(apis.StopTunnel)
}

func (a *TunnelRest) ListTunnels(resp http.ResponseWriter, req *http.Request) {}
func (a *TunnelRest) GetTunnel(resp http.ResponseWriter, req *http.Request) {
	req.GetBody()
}
func (a *TunnelRest) AddTunnel(resp http.ResponseWriter, req *http.Request)    {}
func (a *TunnelRest) UpdateTunnel(resp http.ResponseWriter, req *http.Request) {}
func (a *TunnelRest) RemoveTunnel(resp http.ResponseWriter, req *http.Request) {}
func (a *TunnelRest) StartTunnel(resp http.ResponseWriter, req *http.Request)  {}
func (a *TunnelRest) StopTunnel(resp http.ResponseWriter, req *http.Request)   {}
