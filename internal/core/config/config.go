/*
 * Copyright (C) 2024 by Jason Figge
 */

package config

const (
	Undefined = "<default>"
)

var ( // Build values
	Commit      string
	Version     string
	BuildNumber string
	Release     string
)

var ( // Argument flags
	ConfigFileName string
	Config         *Configuration
	VerboseFlag    bool
	ForcedFlag     bool
	PromptFlag     bool
	CurlFlag       bool
	RawFlag        bool
)

type Configuration struct {
	Hosts   []*Host
	Tunnels []*Tunnel
	Monitor *Monitor
	Web     *Web
}

type Host struct {
	Name       string    `yaml:"name" json:"name"`
	Group      string    `yaml:"group,omitempty" json:"group,omitempty"`
	Address    string    `yaml:"address" json:"address"`
	Username   string    `yaml:"username" json:"username"`
	Identity   string    `yaml:"identity" json:"identity"`
	KnownHosts string    `yaml:"known_hosts" json:"known_hosts"`
	JumpHost   string    `yaml:"jump_host" json:"jump_host"`
	Metadata   *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Tunnel struct {
	Name          string    `yaml:"name" json:"name"`
	Group         string    `yaml:"group,omitempty" json:"group,omitempty"`
	LocalPort     uint16    `yaml:"local_port" json:"local_port"`
	RemoteAddress string    `yaml:"remote_address" json:"remote_address"`
	RemotePort    uint16    `yaml:"remote_port" json:"remote_port"`
	JumpHost      string    `yaml:"jump_host,omitempty" json:"jump_host,omitempty"`
	Metadata      *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Metadata struct {
	Color     string `yaml:"color,omitempty" json:"color,omitempty"`
	Highlight bool   `yaml:"highlight" json:"highlight"`
}

type Monitor struct {
	Compressed bool         `yaml:"compressed,omitempty" json:"compressed,omitempty"`
	Metrics    []string     `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	SortOrder  []*SortOrder `yaml:"sort_order,omitempty" json:"sort_order,omitempty"`
}

type SortOrder struct {
	Metric    string `yaml:"metric" json:"metric"`
	Ascending bool   `yaml:"ascending,omitempty" json:"ascending,omitempty"`
}

type Web struct {
	Port            uint16 `yaml:"port,omitempty" json:"port,omitempty"`
	CertificateFile string `yaml:"certificate_file,omitempty" json:"certificate_file,omitempty"`
	CertificateKey  string `yaml:"certificate_key,omitempty" json:"certificate_key,omitempty"`
}

func NewConfig() *Configuration {
	config := Configuration{
		Hosts:   []*Host{},
		Tunnels: []*Tunnel{},
		Monitor: &Monitor{
			Compressed: false,
			Metrics:    []string{"Id", "Name", "Port", "Rcvd", "Sent", "Open", "Jump", "Last"},
			SortOrder: []*SortOrder{
				{Metric: "Group", Ascending: false},
				{Metric: "Id", Ascending: true},
			},
		},
		Web: &Web{},
	}
	return &config
}

func WriteConfig() {

}

func Validate(c *Configuration) error {
	return nil
}
