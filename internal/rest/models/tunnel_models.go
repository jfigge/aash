/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
	"net/http"

	"us.figge.auto-ssh/internal/core/config"
)

type Tunnel interface {
	ListTunnels(
		ctx context.Context,
		input *ListTunnelInput,
		options ...TunnelOptionFunc,
	) (*ListTunnelOutput, error)
	GetTunnel(
		ctx context.Context,
		input *GetTunnelInput,
		options ...TunnelOptionFunc,
	) (*GetTunnelOutput, error)
	AddTunnel(
		ctx context.Context,
		input *AddTunnelInput,
		options ...TunnelOptionFunc,
	) (*AddTunnelOutput, error)
	UpdateTunnel(
		ctx context.Context,
		input *UpdateTunnelInput,
		options ...TunnelOptionFunc,
	) (*UpdateTunnelOutput, error)
	RemoveTunnel(
		ctx context.Context,
		input *RemoveTunnelInput,
		options ...TunnelOptionFunc,
	) (*RemoveTunnelOutput, error)
	StartTunnel(
		ctx context.Context,
		input *StartTunnelInput,
		options ...TunnelOptionFunc,
	) (*StartTunnelOutput, error)
	StopTunnel(
		ctx context.Context,
		input *StopTunnelInput,
		options ...TunnelOptionFunc,
	) (*StopTunnelOutput, error)
}

type TunnelHeader struct {
	Id     string         `json:"id"`
	Name   string         `json:"name"`
	Status *config.Status `yaml:"status,omitempty" json:"status,omitempty"`
}

type ListTunnelInput struct {
	PaginationInput
	FiltersInput
}

func (i *ListTunnelInput) Vars(req *http.Request) {
	i.PaginationInput.Vars(req)
	i.FiltersInput.Vars(req)
}

type ListTunnelOutput struct {
	Count int             `json:"count"`
	Items []*TunnelHeader `json:"items,omitempty"`
	PaginationOutput
}

type GetTunnelInput struct {
	Id string `json:"id"`
}
type GetTunnelOutput struct {
	config.Tunnel
}

type AddTunnelInput struct{}
type AddTunnelOutput struct{}

type UpdateTunnelInput struct{}
type UpdateTunnelOutput struct{}

type RemoveTunnelInput struct{}
type RemoveTunnelOutput struct{}

type StartTunnelInput struct {
	Id string `json:"id"`
}
type StartTunnelOutput struct {
	Id     string         `json:"id"`
	Status *config.Status `yaml:"status,omitempty" json:"status,omitempty"`
}

type StopTunnelInput struct {
	Id string `json:"id"`
}
type StopTunnelOutput struct {
	Id     string         `json:"id"`
	Status *config.Status `yaml:"status,omitempty" json:"status,omitempty"`
}

type TunnelOptionFunc func(options *TunnelOptions)
type TunnelOptions struct {
	status   bool
	metadata bool
}

func (t *TunnelOptions) Status() bool {
	return t.status
}

func (t *TunnelOptions) Metadata() bool {
	return t.metadata
}

func TunnelOptionStatus(status bool) TunnelOptionFunc {
	return func(options *TunnelOptions) {
		options.status = status
	}
}

func TunnelOptionMetadata(metadata bool) TunnelOptionFunc {
	return func(options *TunnelOptions) {
		options.metadata = metadata
	}
}
