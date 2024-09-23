/*
 * Copyright (C) 2024 by Jason Figge
 */

package core

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"us.figge.auto-ssh/internal/cmd"
	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/core/flag"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays version information about the binary",
	Run: func(cmd *cobra.Command, args []string) {
		err := version(cmd)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(versionCmd)
	flag.AddFlags(versionCmd, flag.Verbose)
}

func version(cmd *cobra.Command) error {
	if cmd.Flag("verbose").Changed {
		format := "%s version %s %s/%s, build %s, commit %s, built %v\n"
		fmt.Printf(format,
			os.Args[0],
			config.Version,
			runtime.GOOS,
			runtime.GOARCH,
			config.BuildNumber,
			config.Commit,
			time.Now().Format(time.DateTime),
		)
	} else {
		fmt.Printf("%s verison %s %s/%s",
			os.Args[0],
			config.Version,
			runtime.GOOS,
			runtime.GOARCH,
		)
	}
	return nil
}
