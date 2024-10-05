/*
 * Copyright (C) 2024 by Jason Figge
 */

package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	managerModels "us.figge.auto-ssh/internal/web/models"
)

const (
	HostName = "hostname"
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
	router.Methods(http.MethodGet).Path("/hosts/{" + HostName + "}").HandlerFunc(apis.GetHost)
	router.Methods(http.MethodPut).Path("/hosts/{" + HostName + "}").HandlerFunc(apis.UpdateHost)
	router.Methods(http.MethodDelete).Path("/hosts/{" + HostName + "}").HandlerFunc(apis.RemoveHost)
}

func (a *HostRest) ListHosts(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.ListHostInput{}
	if req.Body != http.NoBody {
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		input.Validate()
	}
	output, err := a.manager.List(req.Context(), input, extractHostOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
		return
	}
	handleOutputResponse(resp, output)
}

func (a *HostRest) GetHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.GetHostInput{}
	if req.Body != http.NoBody {
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
		}
	}
	output, err := a.manager.Get(req.Context(), input, extractHostOptions(req)...)
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
	_, err = a.manager.Add(req.Context(), input, extractHostOptions(req)...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("AddHost: " + hostName)))
}

func (a *HostRest) UpdateHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.UpdateHostInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.Update(req.Context(), input, extractHostOptions(req)...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("UpdateHost: " + hostName)))
}

func (a *HostRest) RemoveHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.RemoveHostInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	_, err = a.manager.Remove(req.Context(), input, extractHostOptions(req)...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("RemoveHost: " + hostName)))
}

func extractHostOptions(req *http.Request) []managerModels.HostOptionFunc {
	var options []managerModels.HostOptionFunc
	return options
}
