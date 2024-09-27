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
	FileName    string
	C           *Configuration
	VerboseFlag bool
	ForcedFlag  bool
	PromptFlag  bool
	CurlFlag    bool
	RawFlag     bool
)

type Configuration struct {
	Hosts   []*Host   `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	Tunnels []*Tunnel `yaml:"tunnels,omitempty" json:"tunnels,omitempty"`
	Monitor *Monitor  `yaml:"monitor,omitempty" json:"monitor,omitempty"`
	Web     *Web      `yaml:"web,omitempty" json:"web,omitempty"`
}

type Host struct {
	Name       string    `yaml:"name" json:"name"`
	Group      string    `yaml:"group,omitempty" json:"group,omitempty"`
	Address    string    `yaml:"address" json:"address"`
	Username   string    `yaml:"username" json:"username"`
	Identity   string    `yaml:"identity" json:"identity"`
	KnownHosts string    `yaml:"known-hosts" json:"knownHosts"`
	JumpHost   string    `yaml:"jump-host" json:"jumpHost"`
	Metadata   *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Tunnel struct {
	Name          string    `yaml:"name" json:"name"`
	Group         string    `yaml:"group,omitempty" json:"group,omitempty"`
	LocalPort     uint16    `yaml:"local-port" json:"localPort"`
	RemoteAddress string    `yaml:"remote-address" json:"remoteAddress"`
	RemotePort    uint16    `yaml:"remote-port" json:"remotePort"`
	JumpHost      string    `yaml:"jump-host,omitempty" json:"jumpHost,omitempty"`
	Metadata      *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Metadata struct {
	Color     string `yaml:"color,omitempty" json:"color,omitempty"`
	Highlight bool   `yaml:"highlight" json:"highlight"`
}

type Monitor struct {
	Compressed bool         `yaml:"compressed,omitempty" json:"compressed,omitempty"`
	Metrics    []string     `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	SortOrder  []*SortOrder `yaml:"sort-order,omitempty" json:"sortOrder,omitempty"`
}

type SortOrder struct {
	Metric    string `yaml:"metric" json:"metric"`
	Ascending bool   `yaml:"ascending,omitempty" json:"ascending,omitempty"`
}

type Web struct {
	Address         string `yaml:"address" json:"address"`
	Port            int16  `yaml:"port,omitempty" json:"port,omitempty"`
	CertificateFile string `yaml:"certificate-file,omitempty" json:"certificate-file,omitempty"`
	CertificateKey  string `yaml:"certificate-key,omitempty" json:"certificate-key,omitempty"`
	KeyPassphrase   string `yaml:"key-passphrase" json:"key-passphrase"`
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

func (c *Configuration) WriteConfig() {

}
func (c *Configuration) Validate() Validations {
	return NewValidations()
}

func (w *Web) Merge(in *Web) *Web {
	out := *w
	if out.Port == 0 {
		out.Port = in.Port
	}
	if out.Address == "" {
		out.Address = in.Address
	}
	if out.CertificateFile == "" {
		out.CertificateFile = in.CertificateFile
	}
	if out.CertificateKey == "" {
		out.CertificateKey = in.CertificateKey
	}
	if out.KeyPassphrase == "" {
		out.KeyPassphrase = in.KeyPassphrase
	}
	return &out
}
