/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"us.figge.auto-ssh/internal/core/utils"
)

type Pagination struct {
	More     *string `json:"more,omitempty"`
	PageSize int     `json:"pageSize,omitempty"`
	Page     int     `json:"page,omitempty"`
}

func (p *Pagination) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 1
	} else if p.PageSize > 1000 {
		p.PageSize = 1000
	}
}

func Page[S any](items []*S, p *Pagination) ([]*S, *string) {
	start := (p.Page - 1) * p.PageSize
	end := p.Page * p.PageSize
	if start > len(items) {
		return []*S{}, utils.Ptr("")
	}
	if end > len(items) {
		end = len(items)
	}
	if end < len(items) {

	}
	return items[start:end], nil
}

type Filter struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}
