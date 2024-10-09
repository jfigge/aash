/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	filtersRegEx = regexp.MustCompile(`(?i)key=([^,]*),value[s]?=(.*)`)
)

type PaginationInput struct {
	More       *string `json:"more,omitempty"`
	MaxResults int     `json:"maxResults,omitempty"`
}

type PaginationOutput struct {
	More *string `json:"more,omitempty"`
}

func (p *PaginationInput) Validate() {
	if p.MaxResults < 1 {
		p.MaxResults = 100
	} else if p.MaxResults > 1000 {
		p.MaxResults = 1000
	}
}

func (p *PaginationInput) Vars(req *http.Request) {
	vs := req.URL.Query()
	if more := vs.Get("more"); more != "" {
		p.More = &more
	}
	maxResults := vs.Get("maxResults")
	if maxResults == "" {
		return
	}
	p.MaxResults, _ = strconv.Atoi(maxResults)
}

type Filter struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type FiltersInput struct {
	Filters []*Filter `json:"filters,omitempty"`
}

func (f *FiltersInput) Vars(req *http.Request) {
	vs := req.URL.Query()
	filters, ok := vs["filters"]
	if !ok {
		return
	}
	for i, filter := range filters {
		kv := filtersRegEx.FindStringSubmatch(filter)
		if len(kv) != 3 {
			continue
		}
		f.Filters = append(f.Filters, &Filter{
			Key:    kv[1],
			Values: strings.Split(kv[2], ";"),
		})
		if i == 9 {
			break
		}
	}
}
