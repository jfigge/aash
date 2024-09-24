/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
)

type HostOptionFunc func(options *HostOptions)
type HostOptions struct{}
type Host interface {
	List(
		ctx context.Context,
		input *ListHostInput,
		options ...HostOptionFunc,
	) (*ListHostOutput, error)
	Get(
		ctx context.Context,
		input *GetHostInput,
		options ...HostOptionFunc,
	) (*GetHostOutput, error)
	Add(
		ctx context.Context,
		input *AddHostInput,
		options ...HostOptionFunc,
	) (*AddHostOutput, error)
	Update(
		ctx context.Context,
		input *UpdateHostInput,
		options ...HostOptionFunc,
	) (*UpdateHostOutput, error)
	Remove(
		ctx context.Context,
		input *RemoveHostInput,
		options ...HostOptionFunc,
	) (*RemoveHostOutput, error)
}

type ListHostInput struct{}
type ListHostOutput struct{}

type GetHostInput struct{}
type GetHostOutput struct{}

type AddHostInput struct{}
type AddHostOutput struct{}

type UpdateHostInput struct{}
type UpdateHostOutput struct{}

type RemoveHostInput struct{}
type RemoveHostOutput struct{}
