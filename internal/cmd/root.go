/*
 * Copyright (C) 2024 by Jason Figge
 */

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/core/flag"
	engine2 "us.figge.auto-ssh/internal/resources/engine"
	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/rest"
)

const (
	envVarPrefix = "ASH"
)

var (
	ctx             context.Context
	cancel          context.CancelFunc
	server          *rest.Server
	hosts           engineModels.HostEngine
	tunnels         engineModels.TunnelEngine
	stats           engineModels.Stats
	configFilenames = []string{
		".auto-ssh.yaml", ".auto-ssh.yml", ".auto-ssh.json",
		"/auth-ssh/config.yaml", "/auth-ssh/config.yml", "/auth-ssh/config.json"}
)

var RootCmd = &cobra.Command{
	Use:   "ash",
	Short: "auto-ssh command line interface",
	Long:  `A command line for establishing and managing automatic ssh tunneling`,
	Run: func(cmd *cobra.Command, args []string) {
		startEngines()
		startServer()
		startApplication()
	},
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initContext, initConfig)
	flag.AddFlags(RootCmd, rest.Flags, flag.Core)
}

func initConfig() {
	if err := initConfigE(); err != nil {
		fmt.Printf("Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}
}
func initConfigE() error {
	// Locate the configuration file, if one was provided, or search for one
	// in the users home directory or the current directory (current first)
	var bs []byte
	var err error
	var paths []string

	if config.FileName != "" {
		paths = append(paths, config.FileName)
	} else {
		var pwd, home string
		// Fine current directory
		pwd, err = os.Getwd()
		if err != nil {
			pwd = "."
		}
		paths = append(paths, pwd)

		// Find home directory.
		home, err = os.UserHomeDir()
		cobra.CheckErr(err)
		paths = append(paths, home)

		// Etc dir
		if runtime.GOOS != "windows" {
			paths = append(paths, "/etc")
		}
	}

	config.C = config.NewConfig()
	for _, path := range paths {
		for _, filename := range configFilenames {
			config.FileName = filepath.Join(path, filename)
			bs, err = os.ReadFile(config.FileName)
			if err == nil && len(bs) > 0 {
				fmt.Printf("Loading config from %s\n", config.FileName)
				err = yaml.Unmarshal(bs, config.C)
				return err
			}
		}
	}
	fmt.Printf("No config file found.  Setting defaults")
	return nil
}

func initContext() {
	ctx, cancel = context.WithCancel(context.Background())
}

func startEngines() {
	if err := startEnginesE(); err != nil {
		fmt.Printf("failed to start engines: %v\n", err)
		os.Exit(1)
	}
}
func startEnginesE() error {
	var ok bool
	if hosts, ok = engine2.NewHostEngine(ctx, config.C.Hosts); !ok {
		return fmt.Errorf("invalid hosts")
	}
	if tunnels, ok = engine2.NewTunnelEngine(ctx, hosts, config.C.Tunnels); !ok {
		return fmt.Errorf("invalid hosts")
	}
	stats = engine2.NewStatsEngine()
	return nil
}

func startServer() {
	if err := startServerE(); err != nil {
		fmt.Printf("failed to start server: %v\n", err)
		os.Exit(1)
	}
}
func startServerE() error {
	var err error
	server, err = rest.NewServer(ctx, config.C.Web, hosts, tunnels)
	if err != nil {
		return err
	}
	return nil
}

func startApplication() {
	cleanup := func() {
		server.Shutdown()
		cancel()
	}

	stats.StartStatsTunnel(ctx, config.C.Monitor.StatsPort)
	wg, ok := tunnels.StartTunnels(ctx)
	if !ok {
		cleanup()
		fmt.Printf("Failed to successful start tunnels")
		os.Exit(1)
	}

	go func() {
		// Pressing Ctrl+C signals all threads to end. This in turn causes the below wg.Wait() to end
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Printf("\nsystem-service: received signal. Shutting down\n")
	}()
	wg.Wait()
}
