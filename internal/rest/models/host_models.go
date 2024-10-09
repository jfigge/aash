/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
	"net/http"

	"us.figge.auto-ssh/internal/core/config"
)

type HostOptionFunc func(options *HostOptions)
type HostOptions struct{}
type Host interface {
	ListHosts(
		ctx context.Context,
		input *ListHostInput,
		options ...HostOptionFunc,
	) (*ListHostOutput, error)
	GetHost(
		ctx context.Context,
		input *GetHostInput,
		options ...HostOptionFunc,
	) (*GetHostOutput, error)
	AddHost(
		ctx context.Context,
		input *AddHostInput,
		options ...HostOptionFunc,
	) (*AddHostOutput, error)
	UpdateHost(
		ctx context.Context,
		input *UpdateHostInput,
		options ...HostOptionFunc,
	) (*UpdateHostOutput, error)
	RemoveHost(
		ctx context.Context,
		input *RemoveHostInput,
		options ...HostOptionFunc,
	) (*RemoveHostOutput, error)
	ListKnownHosts(
		ctx context.Context,
		input *ListKnownHostsInput,
		options ...HostOptionFunc,
	) (*ListKnownHostsOutput, error)
}

type HostHeader struct {
	Id      string `yaml:"id" json:"id"`
	Name    string `yaml:"name" json:"name"`
	Valid   bool   `yaml:"valid" json:"valid"`
	Running bool   `yaml:"running" json:"running"`
}

type KnownHost struct {
	File string `yaml:"file" json:"file"`
}

type ListHostInput struct {
	PaginationInput
	FiltersInput
}

func (i *ListHostInput) Vars(req *http.Request) {
	i.PaginationInput.Vars(req)
	i.FiltersInput.Vars(req)
}

type ListHostOutput struct {
	Count int           `json:"count"`
	Items []*HostHeader `json:"items,omitempty"`
	PaginationOutput
}

type GetHostInput struct {
	Id string `json:"id"`
}
type GetHostOutput struct {
	config.Host
}

type AddHostInput struct {
	Name       string           `yaml:"name" json:"name"`
	Group      string           `yaml:"group,omitempty" json:"group,omitempty"`
	Address    string           `yaml:"address" json:"address"`
	Username   string           `yaml:"username" json:"username"`
	Identity   string           `yaml:"identity" json:"identity"`
	KnownHosts string           `yaml:"known-hosts" json:"knownHosts"`
	JumpHost   string           `yaml:"jump-host" json:"jumpHost"`
	Metadata   *config.Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}
type AddHostOutput struct{}

type UpdateHostInput struct{}
type UpdateHostOutput struct{}

type RemoveHostInput struct{}
type RemoveHostOutput struct{}

type ListKnownHostsInput struct {
	PaginationInput
}
type ListKnownHostsOutput struct {
	Count int          `json:"count"`
	Items []*KnownHost `json:"items,omitempty"`
	PaginationOutput
}

func (i *ListKnownHostsInput) Vars(req *http.Request) {
	i.PaginationInput.Vars(req)
}
