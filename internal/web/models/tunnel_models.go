/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"

	"us.figge.auto-ssh/internal/core/config"
)

type TunnelOptionFunc func(options *TunnelOptions)
type TunnelOptions struct{}
type Tunnel interface {
	List(
		ctx context.Context,
		input *ListTunnelInput,
		options ...TunnelOptionFunc,
	) (*ListTunnelOutput, error)
	Get(
		ctx context.Context,
		input *GetTunnelInput,
		options ...TunnelOptionFunc,
	) (*GetTunnelOutput, error)
	Add(
		ctx context.Context,
		input *AddTunnelInput,
		options ...TunnelOptionFunc,
	) (*AddTunnelOutput, error)
	Update(
		ctx context.Context,
		input *UpdateTunnelInput,
		options ...TunnelOptionFunc,
	) (*UpdateTunnelOutput, error)
	Remove(
		ctx context.Context,
		input *RemoveTunnelInput,
		options ...TunnelOptionFunc,
	) (*RemoveTunnelOutput, error)
	Start(
		ctx context.Context,
		input *StartTunnelInput,
		options ...TunnelOptionFunc,
	) (*StartTunnelOutput, error)
	Stop(
		ctx context.Context,
		input *StopTunnelInput,
		options ...TunnelOptionFunc,
	) (*StopTunnelOutput, error)
}

type ListTunnelInput struct {
	*Pagination
	Filter []*Filter `json:"filter,omitempty"`
}
type ListTunnelOutput struct {
	Count int              `json:"count"`
	Items []*config.Tunnel `json:"items,omitempty"`
	More  *string          `json:"more,omitempty"`
}

type GetTunnelInput struct{}
type GetTunnelOutput struct{}

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
