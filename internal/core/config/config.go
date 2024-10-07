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
	Id         string    `yaml:"id" json:"id"`
	Name       string    `yaml:"name" json:"name"`
	Remote     *Address  `yaml:"remote" json:"remove"`
	Username   string    `yaml:"username" json:"username"`
	Identity   string    `yaml:"identity" json:"identity"`
	KnownHosts string    `yaml:"knownHosts" json:"knownHosts"`
	JumpHost   string    `yaml:"jumpHost" json:"jumpHost"`
	Metadata   *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Tunnel struct {
	Id       string    `yaml:"id" json:"id"`
	Name     string    `yaml:"name" json:"name"`
	Local    *Address  `yaml:"local" json:"local"`
	Remote   *Address  `yaml:"remote" json:"remote"`
	Host     string    `yaml:"host,omitempty" json:"host,omitempty"`
	Metadata *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Metadata struct {
	Tags      []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Color     string   `yaml:"color,omitempty" json:"color,omitempty"`
	Highlight string   `yaml:"highlight" json:"highlight"`
}

type Monitor struct {
	Color      *Color       `yaml:"color"           json:"color"`
	StatsPort  int          `yaml:"statsPort" json:"statsPort"`
	Compressed bool         `yaml:"compressed,omitempty" json:"compressed,omitempty"`
	Metrics    []string     `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	SortOrder  []*SortOrder `yaml:"sortOrder,omitempty" json:"sortOrder,omitempty"`
	Units      string       `yaml:"units" json:"units"`
}

type Color struct {
	Header  string `yaml:"header"      json:"header"`
	Tunnel  string `yaml:"tunnel"      json:"tunnel"`
	MRU     string `yaml:"mru"         json:"mru"`
	Jump    string `yaml:"jump-tunnel" json:"jump-tunnel"`
	Enabled bool   `yaml:"enabled"     json:"enabled"`
}

type SortOrder struct {
	Metric    string `yaml:"metric" json:"metric"`
	Ascending bool   `yaml:"ascending,omitempty" json:"ascending,omitempty"`
}

type Web struct {
	Address         string `yaml:"address" json:"address"`
	Port            int16  `yaml:"port,omitempty" json:"port,omitempty"`
	CertificateFile string `yaml:"certificateFile,omitempty" json:"certificateFile,omitempty"`
	CertificateKey  string `yaml:"certificateKey,omitempty" json:"certificateKey,omitempty"`
	KeyPassphrase   string `yaml:"keyPassphrase,omitempty" json:"keyPassphrase,omitempty"`
}

func NewConfig() *Configuration {
	config := Configuration{
		Hosts:   []*Host{},
		Tunnels: []*Tunnel{},
		Monitor: &Monitor{
			StatsPort:  2663,
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
