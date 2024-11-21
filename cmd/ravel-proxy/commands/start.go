package commands

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/proxy"
	"github.com/valyentdev/ravel/proxy/edge"
	"github.com/valyentdev/ravel/proxy/local"
	"github.com/valyentdev/ravel/proxy/server"
)

func newStartCmd() *cobra.Command {
	var configPath string
	var mode string

	edge := cobra.Command{
		Use:   "start",
		Short: "Start the proxy",

		RunE: func(cmd *cobra.Command, args []string) error {
			if mode != string(proxy.Edge) && mode != string(proxy.Local) {
				return nil
			}

			return runStart(proxy.Mode(mode), configPath)
		},
	}

	edge.Flags().StringVarP(&mode, "mode", "m", "", "Mode of the proxy (edge, local)")
	edge.MarkFlagRequired("mode")
	edge.Flags().StringVarP(&configPath, "config", "c", "/etc/ravel/proxy.toml", "Path to the config file")

	return &edge
}

func runStart(mode proxy.Mode, configPath string) error {
	config, err := proxy.ReadConfigFile(configPath)
	if err != nil {
		return err
	}

	switch mode {
	case proxy.Edge:
		return runEdge(config)
	case proxy.Local:
		return runLocal(config)
	}

	return nil
}

func runEdge(config *proxy.Config) error {
	edgeProxy := edge.NewRavelProxy(config)
	edgeProxy.Start()

	s := server.NewServer(edgeProxy.Handle, config.Local.Address)

	err := s.ListenAndServeTLS(config.Edge.TLS.CertFile, config.Edge.TLS.KeyFile)
	if err != nil {
		return err
	}

	return nil
}

func runLocal(config *proxy.Config) error {
	proxy := local.NewProxy(config)
	proxy.Start()

	s := server.NewServer(proxy.ServeHTTP, config.Local.Address)

	err := s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
