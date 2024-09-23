/*
 * Copyright (C) 2024 by Jason Figge
 */

package flag

import (
	"github.com/spf13/cobra"
	"us.figge.auto-ssh/internal/core/config"
)

func AddFlags(cmd *cobra.Command, flags ...func(cmd *cobra.Command)) {
	for _, flag := range flags {
		flag(cmd)
	}
}

func Raw(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&config.RawFlag, "raw", "r", false, "prints and unfiltered response")
}

func Curl(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&config.CurlFlag, "curl", false, "print a curl command for the rest call executed")
}

func Config(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&config.FileName, "config", "c", "", "optional configuration file")
}

func Prompt(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&config.PromptFlag, "prompt", "w", false, "prompt for missing information")
}

func Force(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&config.ForcedFlag, "force", "f", false, "force without confirmation or validation")
}

func Verbose(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&config.VerboseFlag, "verbose", "v", false, "displays supplemental information")
}

// Rest adds: curl, raw raw
func Rest(cmd *cobra.Command) {
	Curl(cmd)
	Raw(cmd)
}

// Default adds: config, auth, rest
func Default(cmd *cobra.Command) {
	Config(cmd)
	Rest(cmd)
}

// Core adds: Config Verbose Prompt
func Core(cmd *cobra.Command) {
	Config(cmd)
	Verbose(cmd)
	Prompt(cmd)
}
