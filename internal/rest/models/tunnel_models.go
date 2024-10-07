/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
	"net/http"

	"us.figge.auto-ssh/internal/core/config"
)

type TunnelOptionFunc func(options *TunnelOptions)
type TunnelOptions struct{}
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
	Id    string `json:"id"`
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
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

type StartTunnelInput struct{}
type StartTunnelOutput struct{}

type StopTunnelInput struct{}
type StopTunnelOutput struct{}
