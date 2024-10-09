/*
 * Copyright (C) 2024 by Jason Figge
 */

package models

import (
	"context"
)

type Metadata interface {
	ListStates(
		ctx context.Context,
		options ...MetadataOptionFunc,
	) (*ListMetadataStatesOutput, error)
	ListTags(
		ctx context.Context,
		input *ListMetadataTagsInput,
		options ...MetadataOptionFunc,
	) (*ListMetadataTagsOutput, error)
}

type ListMetadataStatesOutput struct {
	States []string `json:"states"`
}

type ListMetadataTagsInput struct {
	FiltersInput
}

type ListMetadataTagsOutput struct {
	Tags []string `json:"tags"`
}

type MetadataOptionFunc func(options *MetadataOptions)
type MetadataOptions struct {
}
