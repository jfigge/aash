/*
 * Copyright (C) 2024 by Jason Figge
 */

package rest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"us.figge.auto-ssh/internal/core/config"
	engineModels "us.figge.auto-ssh/internal/resources/models"
	"us.figge.auto-ssh/internal/rest/endpoints"
	"us.figge.auto-ssh/internal/rest/managers"
	managerModels "us.figge.auto-ssh/internal/rest/models"
)

var (
	cliArgs = &config.Web{}
)

type Server struct {
	wg            *sync.WaitGroup
	webCfg        *config.Web
	httpServer    *http.Server
	hostManager   managerModels.Host
	tunnelManager managerModels.Tunnel
}

func NewServer(
	ctx context.Context,
	web *config.Web,
	hosts engineModels.HostEngine,
	tunnels engineModels.TunnelEngine,
	wg *sync.WaitGroup,
) (*Server, error) {
	s := &Server{
		webCfg: cliArgs.Merge(web),
		wg:     wg,
	}
	v := s.Validate()
	err := v.Output(fmt.Errorf("failed to validate server configuration"))
	if err != nil {
		return nil, err
	}

	hostMgr, tunnelMgr := s.startManagers(ctx, hosts, tunnels)
	routers := s.startHandlers(ctx, hostMgr, tunnelMgr)
	err = s.Serve(ctx, routers)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func Flags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cliArgs.Address, "address", "A", "0.0.0.0", "Address for aut-ssh API server. Default is 0.0.0.0")
	cmd.Flags().Int16VarP(&cliArgs.Port, "port", "P", 0, "port for auto-ssh API server. Zero port disables server")
	cmd.Flags().StringVar(&cliArgs.CertificateFile, "certificate-file", "", "Certificate required to place aut-ssh in https mode")
	cmd.Flags().StringVar(&cliArgs.CertificateKey, "certificate-key", "", "Certificate private key required to place aut-ssh in https mode")
	cmd.Flags().StringVar(&cliArgs.KeyPassphrase, "passphrase", "", "passphrase used to decrypt certificate key.  See -w to prompt")
}

// routes map[string]http.Handler
func (s *Server) Validate() config.Validations {
	v := config.NewValidations()
	if s.webCfg.Port != 0 {
		s.validateWebAddress(&v)
		s.validatePort(&v)
		s.validateCertFile(&v)
		s.validateCertKey(&v)
	} else {
		v.Infof("web server disabled. web.port=0")
	}
	return v
}
func (s *Server) validateWebAddress(v *config.Validations) {
	// Prepare format of ip address
	if s.webCfg.Address == "" {
		s.webCfg.Address = "0.0.0.0"
	} else if ip := net.ParseIP(s.webCfg.Address); ip != nil {
		s.webCfg.Address = ip.String()
	} else if addrs, err := net.LookupHost(s.webCfg.Address); err != nil {
		v.Errorf("web.address: %v", err)
		return
	} else if len(addrs) > 0 {
		s.webCfg.Address = addrs[0]
	}

	serverIP := net.ParseIP(s.webCfg.Address)
	if serverIP.IsUnspecified() || serverIP.IsLoopback() {
		return
	}
	// check ip address is local
	interfaces, err := net.Interfaces()
	if err != nil {
		v.Errorf("failed to retrieve hosts interfaces: %v", err)
		return
	}
	for _, i := range interfaces {
		addresses, _ := i.Addrs()
		for _, address := range addresses {
			ip, _, _ := net.ParseCIDR(address.String())
			if serverIP.Equal(ip) {
				return
			}
		}
	}
	v.Errorf("web.address must be a valid address on the host")
}
func (s *Server) validatePort(v *config.Validations) {
	if s.webCfg.Port < 0 {
		v.Errorf("web.port cannot be negative")
	} else {
		address := fmt.Sprintf("%s:%d", s.webCfg.Address, s.webCfg.Port)
		ln, err := net.Listen("tcp", address)
		if err != nil {
			v.Errorf("web.port is already in use [%s]", address)
		} else if ln != nil {
			_ = ln.Close()
		}
	}
}
func (s *Server) validateCertFile(v *config.Validations) {
	if s.webCfg.CertificateFile == "" {
		v.Warnf("web.certificate_file not set.  auto.ssh web server will use http")
		return
	}
	if fi, err := os.Stat(s.webCfg.CertificateFile); err != nil {
		v.Errorf("Unable to retrieve stats on web.certificate_file: %v", err)
		return
	} else if fi.IsDir() {
		v.Errorf("web.certificate_file points to a directory, not a certificate")
	}
	if _, err := os.ReadFile(s.webCfg.CertificateFile); err != nil {
		v.Errorf("web.certificate_file cannot be read: %v", err)
	}
}
func (s *Server) validateCertKey(v *config.Validations) {
	if s.webCfg.CertificateFile == "" {
		return
	} else if s.webCfg.CertificateKey == "" {
		v.Errorf("web.certificate_key must be specified if web.certificate_file is set")
		return
	}
	if fi, err := os.Stat(s.webCfg.CertificateKey); err != nil {
		v.Errorf("Unable to retrieve stats on web.certificate_key: %v", err)
		return
	} else if fi.IsDir() {
		v.Errorf("web.certificate_key points to a directory, not a private key")
	}
	if _, err := os.ReadFile(s.webCfg.CertificateKey); err != nil {
		v.Errorf("web.certificate_key cannot be read: %v", err)
	}
}

func (s *Server) startManagers(
	ctx context.Context, hosts engineModels.HostEngine, tunnels engineModels.TunnelEngine,
) (managerModels.Host, managerModels.Tunnel) {
	hostManager, tunnelManager, err := s.startManagersE(ctx, hosts, tunnels)
	if err != nil {
		fmt.Printf("failed to start managers: %v\n", err)
		os.Exit(1)
	}
	return hostManager, tunnelManager
}
func (s *Server) startManagersE(
	ctx context.Context, hosts engineModels.HostEngine, tunnels engineModels.TunnelEngine,
) (hostManager managerModels.Host, tunnelManager managerModels.Tunnel, err error) {
	hostManager, err = managers.NewHostManager(ctx, hosts)
	if err != nil {
		return
	}
	tunnelManager, err = managers.NewTunnelManager(ctx, tunnels)
	if err != nil {
		return
	}
	return
}

func (s *Server) startHandlers(
	ctx context.Context, hostManager managerModels.Host, tunnelManager managerModels.Tunnel,
) *mux.Router {
	routes := mux.NewRouter()
	endpoints.NewHostRest(ctx, hostManager, routes)
	endpoints.NewTunnelRest(ctx, tunnelManager, routes)
	return routes
}

func (s *Server) Serve(ctx context.Context, routes *mux.Router) error {
	s.wg.Add(1)
	listenAddress := fmt.Sprintf("%s:%d", s.webCfg.Address, s.webCfg.Port)
	//nolint: gosec
	s.httpServer = &http.Server{
		Handler: routes,
	}
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	if s.webCfg.CertificateFile != "" {
		certFile := s.webCfg.CertificateFile
		keyFile := s.webCfg.CertificateKey
		go s.serveHTTPS(ln, listenAddress, certFile, keyFile)
	} else {
		go s.serveHTTP(ln, listenAddress)
	}
	return nil
}
func (s *Server) serveHTTPS(ln net.Listener, listenAddress, certFile, keyFile string) {
	fmt.Printf("Listening on https -> %s\n", listenAddress)
	err := s.httpServer.ServeTLS(ln, certFile, keyFile)
	if err != nil {
		fmt.Printf("web server has shut down: %v\n", err)
	}
}
func (s *Server) serveHTTP(ln net.Listener, listenAddress string) {
	fmt.Printf("Listening on http -> %s\n", listenAddress)
	err := s.httpServer.Serve(ln)
	if err != nil {
		fmt.Printf("web server has shut down: %v\n", err)
	}
}
func (s *Server) Shutdown() {
	if s.httpServer != nil {
		err := s.httpServer.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("error shutting down web server: %v", err)
		}
		fmt.Printf("server is shut down\n")
		s.httpServer = nil
		s.wg.Done()
	}
}
