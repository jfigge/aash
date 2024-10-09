/*
 * Copyright (C) 2024 by Jason Figge
 */

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	managerModels "us.figge.auto-ssh/internal/rest/models"
)

type MetadataRest struct {
	manager managerModels.Metadata
}

func NewMetadataRest(ctx context.Context, manager managerModels.Metadata, router *mux.Router) {
	apis := &MetadataRest{
		manager: manager,
	}
	router.Methods(http.MethodGet, http.MethodPost).Path("/metadata/states").HandlerFunc(apis.States)
	router.Methods(http.MethodGet, http.MethodPost).Path("/metadata/tags").HandlerFunc(apis.Tags)
}

func (m MetadataRest) States(resp http.ResponseWriter, req *http.Request) {
	output, err := m.manager.ListStates(req.Context(), extractMetadataOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
		return
	}
	handleOutputResponse(resp, output)
}

func (m MetadataRest) Tags(resp http.ResponseWriter, req *http.Request) {
	input := &managerModels.ListMetadataTagsInput{}
	if req.Method == http.MethodGet {
		input.Vars(req)
	} else if req.Body != http.NoBody {
		err := json.NewDecoder(req.Body).Decode(input)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	output, err := m.manager.ListTags(req.Context(), input, extractMetadataOptions(req)...)
	if err != nil {
		handleErrorResponse(resp, err)
		return
	}
	handleOutputResponse(resp, output)
}

func extractMetadataOptions(req *http.Request) []managerModels.MetadataOptionFunc {
	var opts []managerModels.MetadataOptionFunc
	return opts
}
