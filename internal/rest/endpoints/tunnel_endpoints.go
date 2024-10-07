/*
 * Copyright (C) 2024 by Jason Figge
 */

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	managerModels "us.figge.auto-ssh/internal/rest/models"
)

type TunnelRest struct {
	manager managerModels.Tunnel
}

func NewTunnelRest(ctx context.Context, manager managerModels.Tunnel, router *mux.Router) {
	apis := &TunnelRest{
		manager: manager,
	}
	router.Methods(http.MethodGet, http.MethodPost).Path("/tunnels").HandlerFunc(apis.ListTunnels)
	router.Methods(http.MethodPost).Path("/tunnels").HandlerFunc(apis.AddTunnel)
	router.Methods(http.MethodGet).Path("/tunnels/{id}").HandlerFunc(apis.GetTunnel)
	router.Methods(http.MethodPut).Path("/tunnels/{id}").HandlerFunc(apis.UpdateTunnel)
	router.Methods(http.MethodDelete).Path("/tunnels/{id}").HandlerFunc(apis.RemoveTunnel)
	router.Methods(http.MethodDelete).Path("/tunnels/{id}/start").HandlerFunc(apis.StartTunnel)
	router.Methods(http.MethodDelete).Path("/tunnels/{id}/stop").HandlerFunc(apis.StopTunnel)
}

func (a *TunnelRest) ListTunnels(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.ListTunnelInput{}
	if req.Method == http.MethodGet {
		input.Vars(req)
	} else if req.Body != http.NoBody {
		err := json.NewDecoder(req.Body).Decode(input)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	input.Validate()
	output, err := a.manager.ListTunnels(req.Context(), input, extractTunnelOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
		return
	}
	handleOutputResponse(resp, output)
}
func (a *TunnelRest) GetTunnel(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.GetTunnelInput{}
	input.Id = mux.Vars(req)[id]
	output, err := a.manager.GetTunnel(req.Context(), input, extractTunnelOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
	}
	handleOutputResponse(resp, output)
}

func (a *TunnelRest) AddTunnel(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.AddTunnelInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.AddTunnel(req.Context(), input, extractTunnelOptions(req)...)
	hostName := mux.Vars(req)[id]
	resp.Write([]byte(fmt.Sprintf("AddTunnel: " + hostName)))
}

func (a *TunnelRest) UpdateTunnel(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.UpdateTunnelInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.UpdateTunnel(req.Context(), input, extractTunnelOptions(req)...)
	hostName := mux.Vars(req)[id]
	resp.Write([]byte(fmt.Sprintf("UpdateTunnel: " + hostName)))
}

func (a *TunnelRest) RemoveTunnel(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.RemoveTunnelInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.RemoveTunnel(req.Context(), input, extractTunnelOptions(req)...)
	hostName := mux.Vars(req)[id]
	resp.Write([]byte(fmt.Sprintf("RemoveTunnel: " + hostName)))
}

func (a *TunnelRest) StartTunnel(resp http.ResponseWriter, req *http.Request) {}
func (a *TunnelRest) StopTunnel(resp http.ResponseWriter, req *http.Request)  {}

func extractTunnelOptions(req *http.Request) []managerModels.TunnelOptionFunc {
	var options []managerModels.TunnelOptionFunc
	return options
}
