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
	router.Methods(http.MethodGet).Path("/hosts").HandlerFunc(apis.ListHosts)
	router.Methods(http.MethodPost).Path("/hosts").HandlerFunc(apis.AddHost)
	router.Methods(http.MethodGet).Path("/hosts/{" + HostName + "}").HandlerFunc(apis.GetHost)
	router.Methods(http.MethodPut).Path("/hosts/{" + HostName + "}").HandlerFunc(apis.UpdateHost)
	router.Methods(http.MethodDelete).Path("/hosts/{" + HostName + "}").HandlerFunc(apis.RemoveHost)
}

func (a *HostRest) ListHosts(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.ListHostInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	options := a.extractOptions(req)
	_, err = a.manager.List(req.Context(), input, options...)
	resp.Write([]byte("ListHosts"))
}

func (a *HostRest) GetHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.GetHostInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	options := a.extractOptions(req)
	_, err = a.manager.Get(req.Context(), input, options...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("GetHost: " + hostName)))
}

func (a *HostRest) AddHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.AddHostInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	options := a.extractOptions(req)
	_, err = a.manager.Add(req.Context(), input, options...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("AddHost: " + hostName)))
}

func (a *HostRest) UpdateHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.UpdateHostInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	options := a.extractOptions(req)
	_, err = a.manager.Update(req.Context(), input, options...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("UpdateHost: " + hostName)))
}

func (a *HostRest) RemoveHost(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.RemoveHostInput{}
	err := json.NewDecoder(req.Body).Decode(input)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
	}
	options := a.extractOptions(req)
	_, err = a.manager.Remove(req.Context(), input, options...)
	hostName := mux.Vars(req)[HostName]
	resp.Write([]byte(fmt.Sprintf("RemoveHost: " + hostName)))
}

func (a *HostRest) extractOptions(req *http.Request) []managerModels.HostOptionFunc {
	var options []managerModels.HostOptionFunc
	return options
}
