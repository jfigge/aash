/*
 * Copyright (C) 2024 by Jason Figge
 */

package host

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"us.figge.auto-ssh/internal/core/config"
)

type hostData struct {
	*config.Host
	lock       sync.Mutex
	valid      bool
	inUse      bool
	referenced bool
	isJumpHost bool
	client     *ssh.Client
	config     *ssh.ClientConfig
}
type Entry struct {
	*hostData
}

func (h *Entry) Id() string {
	return h.hostData.Id
}
func (h *Entry) Name() string {
	return h.hostData.Name
}
func (h *Entry) Remote() *config.Address {
	return h.hostData.Remote
}
func (h *Entry) Username() string {
	return h.hostData.Username
}
func (h *Entry) Passphrase() string {
	return h.hostData.Passphrase
}
func (h *Entry) Identity() string {
	return h.hostData.Identity
}
func (h *Entry) KnownHosts() string {
	return h.hostData.KnownHosts
}
func (h *Entry) JumpHost() string {
	return h.hostData.JumpHost
}
func (h *Entry) Valid() bool {
	return h.hostData.valid
}
func (h *Entry) Metadata() *config.Metadata {
	return h.hostData.Metadata
}
func (h *Entry) Referenced() {
	h.referenced = true
}

//func (h *Entry) IsJumpHost() bool {
//	return h.hostData.isJumpHost
//}
//func (h *Entry) IsHost() bool {
//	return h.hostData.inUse
//}

func (h *Entry) Open() bool {
	//h.lock.Lock()
	//defer h.lock.Unlock()
	return h.open()
}
func (h *Entry) open() bool {
	if h.client == nil {
		var err error
		h.client, err = ssh.Dial("tcp", h.hostData.Remote.String(), h.config)
		if err != nil {
			fmt.Printf("  Error - failed to connect to remote address: %v\n", err)
			return false
		}
	}
	return true
}

func (h *Entry) Dial(address string) (net.Conn, bool) {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.redial(address, false)
}

func (h *Entry) redial(address string, redialing bool) (net.Conn, bool) {
	conn, err := h.client.Dial("tcp", address)
	if err != nil {
		_ = h.client.Close()
		h.client = nil
		if !redialing {
			if h.open() {
				return h.redial(address, true)
			} else {
				return nil, false
			}
		}
		fmt.Printf("  Error - Host (%s) failed to call forward address: %v\n", h.hostData.Name, err)
		return nil, false
	}
	return conn, true
}

func (h *Entry) Validate(
	defaultUsername string,
	identityMap map[string]ssh.Signer,
	hostKeysMap map[string]*HostKeyManager,
) bool {
	warning := false
	h.hostData.Name = strings.TrimSpace(h.hostData.Name)
	if h.hostData.Name == "" {
		fmt.Printf("  Error - host name cannot be blank\n")
		h.valid = false
	}

	h.hostData.Username = strings.TrimSpace(h.hostData.Username)
	if strings.TrimSpace(h.hostData.Username) == "" && config.VerboseFlag {
		fmt.Printf("  Info  - host (%s) will use default username: %s\n", h.hostData.Name, defaultUsername)
		h.hostData.Username = defaultUsername
	}

	h.hostData.KnownHosts = strings.TrimSpace(h.hostData.KnownHosts)
	if h.hostData.KnownHosts == "" {
		fmt.Printf("  Warn  - host (%s) not using a known_hosts file\n", h.hostData.Name)
		warning = true
	} else if _, ok := hostKeysMap[h.hostData.KnownHosts]; !ok {
		if fi, err := os.Stat(h.hostData.KnownHosts); os.IsNotExist(err) {
			fmt.Printf("  Error - host (%s) known_hosts file (%s) cannot be read: file not found\n", h.hostData.Name, h.hostData.KnownHosts)
			h.valid = false
		} else if fi.IsDir() {
			fmt.Printf("  Error - host (%s) known_hosts file (%s) cannot be read: file is a directory\n", h.hostData.Name, h.hostData.KnownHosts)
			h.valid = false
		} else {
			var hkManager *HostKeyManager
			if hkManager, err = NewHostKeyManager(h.hostData.KnownHosts); os.IsPermission(err) {
				fmt.Printf("  Error - host (%s) known_hosts file (%s) cannot be read: permission denied\n", h.hostData.Name, h.hostData.KnownHosts)
				h.valid = false
			} else if err != nil {
				fmt.Printf("  Error - host (%s) known_hosts file (%s) cannot be read: %v\n", h.hostData.Name, h.hostData.KnownHosts, err)
				h.valid = false
			} else {
				hostKeysMap[h.hostData.KnownHosts] = hkManager
			}
		}
	}

	h.hostData.Identity = strings.TrimSpace(h.hostData.Identity)
	if h.hostData.Identity == "" {
		fmt.Printf("  Error - host (%s) missing identity file\n", h.hostData.Name)
		h.valid = false
	}
	if _, ok := identityMap[h.hostData.Identity]; !ok {
		if fi, err := os.Stat(h.hostData.Identity); os.IsNotExist(err) {
			fmt.Printf("  Error - host (%s) identity file (%s) cannot be read: file not found\n", h.hostData.Name, h.hostData.Identity)
			h.valid = false
		} else if fi.IsDir() {
			fmt.Printf("  Error - host (%s) identity file (%s) cannot be read: file is a directory\n", h.hostData.Name, h.hostData.Identity)
			h.valid = false
		} else {
			var key []byte
			key, err = os.ReadFile(h.hostData.Identity)
			if os.IsPermission(err) {
				fmt.Printf("  Error - host (%s) identity file (%s) cannot be read: permission denied\n", h.hostData.Name, h.hostData.Identity)
				h.valid = false
			} else if err != nil {
				fmt.Printf("  Error - host (%s) identity file (%s) cannot be read: %v\n", h.hostData.Name, h.hostData.Identity, err)
				h.valid = false
			} else {
				var signer ssh.Signer
				h.hostData.Passphrase = strings.TrimSpace(h.hostData.Passphrase)
				if h.hostData.Passphrase != "" {
					signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(h.hostData.Passphrase))
				} else {
					signer, err = ssh.ParsePrivateKey(key)
				}
				if err != nil {
					fmt.Printf("  Error - host (%s) identity file (%s) cannot be decode: %v\n", h.hostData.Name, h.hostData.Identity, err)
					h.valid = false
				} else {
					identityMap[h.hostData.Identity] = signer
				}
			}
		}
	}

	if h.hostData.Remote == nil || h.hostData.Remote.IsBlank() {
		fmt.Printf("  Error - host (%s) requires an address\n", h.hostData.Name)
		h.valid = false
	} else if !h.hostData.Remote.Validate("host", h.hostData.Name, "address", h.hostData.JumpHost != "", true) {
		h.valid = false
	}

	if h.hostData.JumpHost != "" {
		if h.hostData.JumpHost == h.hostData.Name {
			fmt.Printf("  Error - host (%s) jump_host cannot reference itself\n", h.hostData.Name)
			h.valid = false
		} else {
			h.hostData.KnownHosts = ""
		}
	}
	h.config = &ssh.ClientConfig{
		User: h.hostData.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(identityMap[h.hostData.Identity]),
		},
		HostKeyCallback: hostKeysMap[h.hostData.KnownHosts].Callback,
	}

	if config.VerboseFlag && h.valid && !warning {
		fmt.Printf("  Info  - host (%s) validated\n", h.hostData.Name)
	}
	return h.valid
}
