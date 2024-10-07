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

type HostRest struct {
	manager managerModels.Host
}

func NewHostRest(ctx context.Context, manager managerModels.Host, router *mux.Router) {
	apis := &HostRest{
		manager: manager,
	}
	router.Methods(http.MethodGet, http.MethodPost).Path("/hosts").HandlerFunc(apis.ListHosts)
	router.Methods(http.MethodPost).Path("/hosts").HandlerFunc(apis.AddHost)
	router.Methods(http.MethodGet).Path("/hosts/known-hosts").HandlerFunc(apis.ListKnownHosts)
	router.Methods(http.MethodGet).Path("/hosts/{id}").HandlerFunc(apis.GetHost)
	router.Methods(http.MethodPut).Path("/hosts/{id}").HandlerFunc(apis.UpdateHost)
	router.Methods(http.MethodDelete).Path("/hosts/{id}").HandlerFunc(apis.RemoveHost)
}

func (a *HostRest) ListHosts(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.ListHostInput{}
	if req.Method == http.MethodGet {
		input.Vars(req)
	} else if req.Body != http.NoBody {
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	input.Validate()
	output, err := a.manager.ListHosts(req.Context(), input, extractHostOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
		return
	}
	handleOutputResponse(resp, output)
}

func (a *HostRest) GetHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.GetHostInput{}
	input.Id = mux.Vars(req)[id]
	output, err := a.manager.GetHost(req.Context(), input, extractHostOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
	}
	handleOutputResponse(resp, output)
}

func (a *HostRest) AddHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.AddHostInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.AddHost(req.Context(), input, extractHostOptions(req)...)
	hostName := mux.Vars(req)[id]
	resp.Write([]byte(fmt.Sprintf("AddHost: " + hostName)))
}

func (a *HostRest) UpdateHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.UpdateHostInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.UpdateHost(req.Context(), input, extractHostOptions(req)...)
	hostName := mux.Vars(req)[id]
	resp.Write([]byte(fmt.Sprintf("UpdateHost: " + hostName)))
}

func (a *HostRest) RemoveHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.RemoveHostInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.RemoveHost(req.Context(), input, extractHostOptions(req)...)
	hostName := mux.Vars(req)[id]
	resp.Write([]byte(fmt.Sprintf("RemoveHost: " + hostName)))
}

func (a *HostRest) ListKnownHosts(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.ListKnownHostsInput{}
	if req.Method == http.MethodGet {
		input.Vars(req)
	}
	input.Validate()
	output, err := a.manager.ListKnownHosts(req.Context(), input, extractHostOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
		return
	}
	handleOutputResponse(resp, output)
}

func extractHostOptions(req *http.Request) []managerModels.HostOptionFunc {
	var options []managerModels.HostOptionFunc
	return options
}
