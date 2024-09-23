/*
 * Copyright (C) 2024 by Jason Figge
 */

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"us.figge.auto-ssh/internal/core/config"
	"us.figge.auto-ssh/internal/core/flag"
	"us.figge.auto-ssh/internal/web"
)

const (
	envVarPrefix = "ASH"
)

var RootCmd = &cobra.Command{
	Use:   "ash",
	Short: "auto-ssh command line interface",
	Long:  `A command line for establishing and managing automatic ssh tunneling`,
	Run: func(cmd *cobra.Command, args []string) {
		if launch() != nil {
			os.Exit(1)
		}
	},
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initRootValidate)
	flag.AddFlags(RootCmd, web.Flags, flag.Core)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := initConfigE(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
}

// initConfigE
func initConfigE() error {
	// Set the tone early and let viper know we're using yaml
	v := viper.New()
	v.SetConfigType("yaml")

	// Initialize viper with a default set of configurations. In our case the
	// configurations are all blank, but this still gives viper an opportunity
	// to see all the names of the parameters.  This will help shortly...
	out, err := yaml.Marshal(config.NewConfig())
	if err != nil {
		return err
	}

	err = v.MergeConfig(bytes.NewReader(out))
	if err != nil {
		return err
	}

	// Locate the configuration file, if one was provided, or search for one
	// in the users home directory or the current directory (current first)
	if config.FileName != "" {
		// Use config file from the flag.
		v.SetConfigFile(config.FileName)
	} else {
		var pwd, home string
		// Fine current directory
		pwd, err = os.Getwd()
		if err != nil {
			pwd = "."
		}
		v.AddConfigPath(pwd)
		if runtime.GOOS != "windows" {
			v.AddConfigPath("/etc")
		}

		// Find home directory.
		home, err = os.UserHomeDir()
		cobra.CheckErr(err)
		v.AddConfigPath(home)

		v.SetConfigName(".auto-ssh")
	}

	// If we found a configuration file then we need to overlay the defaults with the
	// content of the config file.
	err = v.MergeInConfig()
	if err == nil {
		config.FileName = v.ConfigFileUsed()
	} else {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Having a config file is optional, so we only return an error if it's
			// not related to a missing file
			return err
		}
	}

	// Now we overlay the environment variables. This is why we needed to
	// see all the defaults upfront, as now we know what environment variables
	// to look for. Note: All variables will be prefixed with PINGCLI_
	v.SetEnvPrefix(envVarPrefix)
	v.AutomaticEnv()                                            // read in environment variables that match
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "")) // Incase we have a complex config
	v.SetTypeByDefaultValue(true)

	// alias hyphened names to struct element names
	//v.RegisterAlias("KeysPaths", "Keys")
	//v.RegisterAlias("PassPhrases", "Passphrases")

	// Finally, we bind viper configs to cobra flags.  If a flag wasn't specified then
	// we set its value to the viper config value, and when a flag was set then we
	// update viper to match. Really we only care about viper values at this point.
	RootCmd.Flags().VisitAll(
		func(flag *pflag.Flag) {
			envVarSuffix := strings.ToUpper(flag.Name)
			if err = v.BindEnv(flag.Name, fmt.Sprintf("%s_%s", envVarPrefix, envVarSuffix)); err == nil {
				if !flag.Changed && v.IsSet(flag.Name) {
					// Push a config value to the flags
					_ = RootCmd.Flags().Set(flag.Name, fmt.Sprintf("%v", v.Get(flag.Name)))
				} else if flag.Changed || !v.IsSet(flag.Name) {
					// push the flag to the config
					if flag.Value.Type() == "stringSlice" {
						values := strings.Split(flag.Value.String()[1:len(flag.Value.String())-1], ",")
						if v.IsSet(flag.Name) {
							values = merge(values, v.Get(flag.Name).([]interface{}))
						}
						v.Set(flag.Name, values)
					} else {
						v.Set(flag.Name, flag.Value)
					}
				}
			}
		},
	)

	// Now that everything has been layered in the correct order we generate the
	// final configuration and make this available for all to use.
	config.C = &config.Configuration{}
	err = v.Unmarshal(&config.C)
	if err != nil {
		return err
	}

	return nil
}

func merge(array1 []string, array2 []interface{}) []string {
	combined := append([]string{}, array1...)
	for _, tmp := range array2 {
		if tmp == nil {
			continue
		}
		a2 := fmt.Sprintf("%v", tmp)
		found := false
		for _, a1 := range array1 {
			if strings.EqualFold(a1, a2) {
				found = true
				break
			}
		}
		if !found {
			combined = append(combined, a2)
		}
	}
	return combined
}

// initRootValidate reads in config file and ENV variables if set.
func initRootValidate() {
	if err := initRootValidateE(); err != nil {
		fmt.Printf("failed to validate root configuration: %v\n", err)
		os.Exit(1)
	}
}

// initRootValidateE
func initRootValidateE() error {
	validations := config.C.Validate()
	return validations.Output(fmt.Errorf("invalid configuration: %s", config.FileName))
}

func launch() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server, err := web.NewServer(config.C.Web)
	if err != nil {
		return err
	}

	cleanup := func(exitCode int) {
		server.Shutdown()
		cancel()
		os.Exit(exitCode)
	}

	_, err = server.Serve(ctx)
	if err != nil {
		return err
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan
	fmt.Printf("system-service: received signal. Shutting down\n")
	cleanup(0)
	return nil
}
