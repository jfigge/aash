/*
 * Copyright (C) 2024 by Jason Figge
 */

package config

import (
	"fmt"
)

type Validations struct {
	hasErrors bool
	entries   []*ValidationEntry
}
type ValidationEntry struct {
	isError bool
	message string
}

func NewValidations() Validations {
	return Validations{
		hasErrors: false,
		entries:   []*ValidationEntry{},
	}
}
func (v *Validations) HasValidations() bool {
	return len(v.entries) != 0
}
func (v *Validations) HasValidationErrors() bool {
	return v.hasErrors
}
func (v *Validations) Validations() []*ValidationEntry {
	return v.entries
}
func (v *Validations) Errorf(msg string, args ...any) {
	v.hasErrors = true
	v.entries = append(v.entries, &ValidationEntry{isError: true, message: "  Error - " + fmt.Sprintf(msg, args...)})
}

func (v *Validations) Warnf(msg string, args ...any) {
	v.entries = append(v.entries, &ValidationEntry{isError: false, message: "  Warn  - " + fmt.Sprintf(msg, args...)})
}

func (v *Validations) Infof(msg string, args ...any) {
	v.entries = append(v.entries, &ValidationEntry{isError: false, message: "  Info  - " + fmt.Sprintf(msg, args...)})
}

func (v *Validations) Output(returnErr error) error {
	var err error
	if v.HasValidations() {
		if v.HasValidationErrors() {
			err = returnErr
			fmt.Printf("One or more configuration validation errors were generated:\n")
		} else if VerboseFlag {
			fmt.Printf("One or more configuration validation warnings were generated:\n")
		}
		for _, entry := range v.Validations() {
			if entry.IsError() || VerboseFlag {
				fmt.Printf("%s\n", entry.Message())
			}
		}
	}
	return err
}

func (ve *ValidationEntry) IsError() bool {
	return ve.isError
}
func (ve *ValidationEntry) Message() string {
	return ve.message
}
