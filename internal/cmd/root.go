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
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/core/flag"
	"us.figge.auto-ssh/internal/resources/engine/host"
	engineStats "us.figge.auto-ssh/internal/resources/engine/stats"
	engineTunnel "us.figge.auto-ssh/internal/resources/engine/tunnel"
	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/rest"
)

var (
	ctx             context.Context
	cancel          context.CancelFunc
	server          *rest.Server
	hostEngine      engineModels.HostEngineInternal
	tunnelEngine    engineModels.TunnelEngine
	statsEngine     engineModels.StatsEngine
	wg              = &sync.WaitGroup{}
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
	hostEngine = host.NewEngine(ctx, config.C.Hosts)
	tunnelEngine = engineTunnel.NewEngine(ctx, hostEngine, config.C.Tunnels)
	statsEngine = engineStats.NewEngine()
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
	server, err = rest.NewServer(ctx, config.C.Web, hostEngine, tunnelEngine, wg)
	if err != nil {
		return err
	}
	return nil
}

func startApplication() {
	err := statsEngine.StartStatsTunnel(ctx, config.C.Monitor.StatsPort)
	if err != nil {
		return
	}
	tunnelEngine.StartTunnels(ctx, statsEngine, wg)

	go func() {
		// Pressing Ctrl+C signals all threads to end. This in turn causes the below wg.Wait() to end
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Printf("\nsystem-service: received signal. Shutting down\n")
		server.Shutdown()
		cancel()
	}()

	wg.Wait()
	server.Shutdown()
	cancel()
}
