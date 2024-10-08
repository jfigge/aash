/*
 * Copyright (C) 2024 by Jason Figge
 */

package host

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type hostKeyEntry struct {
	hash string
	line int
}

type HostKeyManager struct {
	lock          sync.Mutex
	knownHostFile string
	knownKeys     map[string]map[string]hostKeyEntry
	lines         int
}

var (
	InsecureHostKey = &HostKeyManager{
		knownHostFile: "",
	}
)

func NewHostKeyManager(knownHostFile string) (*HostKeyManager, error) {
	bs, err := os.ReadFile(knownHostFile)
	if err != nil {
		return nil, err
	}
	var hs []string
	var pk ssh.PublicKey

	knownKeys := make(map[string]map[string]hostKeyEntry)
	line := 1
	for _, hs, pk, _, bs, err = ssh.ParseKnownHosts(bs); err == nil; _, hs, pk, _, bs, err = ssh.ParseKnownHosts(bs) {
		key := hostKeyEntry{hash: base64.StdEncoding.EncodeToString(pk.Marshal())}
		for _, h := range hs {
			if types, ok := knownKeys[h]; !ok {
				knownKeys[h] = map[string]hostKeyEntry{pk.Type(): key}
			} else if knownKey, ok2 := types[pk.Type()]; !ok2 {
				types[pk.Type()] = key
			} else if knownKey.hash == key.hash {
				fmt.Printf("  Info  - known_hosts (%s) duplicate entries on lines %d and %d\n", knownHostFile, key.line, line)
			} else {
				return nil, fmt.Errorf("known_hosts (%s) inconsistent entries on lines %d and %d\n", knownHostFile, key.line, line)
			}
		}
		line++
	}
	if err.Error() != "EOF" {
		return nil, fmt.Errorf("%w, line %d", err, line)
	}
	return &HostKeyManager{
		knownHostFile: knownHostFile,
		knownKeys:     knownKeys,
		lines:         line,
	}, nil
}

func (h *HostKeyManager) Callback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	if h.knownKeys == nil {
		return nil
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	ip := knownhosts.Normalize(hostname)
	hash := base64.StdEncoding.EncodeToString(key.Marshal())
	types, ok := h.knownKeys[ip]
	if !ok {
		err := h.appendHostKey(hostname, key)
		if err == nil {
			h.lines++
			h.knownKeys[ip] = map[string]hostKeyEntry{key.Type(): {hash: hash, line: h.lines}}
		}
		return err
	}
	knownKey, ok2 := types[key.Type()]
	if !ok2 {
		err := h.appendHostKey(hostname, key)
		if err == nil {
			h.lines++
			types[key.Type()] = hostKeyEntry{hash: hash, line: h.lines}
		}
		return err
	}
	if knownKey.hash == hash {
		return nil
	}
	return fmt.Errorf("the authenticity of host '%s' can't be established", ip)
}

func (h *HostKeyManager) appendHostKey(hostname string, key ssh.PublicKey) error {
	if h.knownHostFile == "" {
		return nil
	}
	ip := knownhosts.Normalize(hostname)
	fmt.Printf("Warning: Permanently added '%s' (%s) to the list of known hosts.\n", ip, key.Type())
	line := fmt.Sprintf("%s %s %s", ip, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))

	f, err := os.OpenFile(h.knownHostFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("  Error - failed to append known host to %s: %v\n", h.knownHostFile, err)
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err = f.WriteString(line); err != nil {
		fmt.Printf("  Error - failed to write known host to %s: %v\n", h.knownHostFile, err)
		return err
	}
	return nil
}
