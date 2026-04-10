package ctl

import (
	"fmt"

	"github.com/alexisbouchez/ravel/pkg/ravelctl"
	"github.com/spf13/cobra"
)

var (
	apiURL    string
	token     string
	namespace string
	outputFmt string
)

func NewCtlCmd() *cobra.Command {
	ctlCmd := &cobra.Command{
		Use:   "ctl",
		Short: "Control ravel clusters via API",
		Long:  `Commands to manage ravel resources via the API server.`,
	}

	ctlCmd.PersistentFlags().StringVar(&apiURL, "api", "", "API server URL (default from ~/.ravel/config.json)")
	ctlCmd.PersistentFlags().StringVar(&token, "token", "", "API token (default from ~/.ravel/config.json)")
	ctlCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to operate in")
	ctlCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format: table, json")

	ctlCmd.AddCommand(newMachinesCmd())
	ctlCmd.AddCommand(newFleetsCmd())
	ctlCmd.AddCommand(newNamespacesCmd())
	ctlCmd.AddCommand(newGatewaysCmd())
	ctlCmd.AddCommand(newNodesCmd())
	ctlCmd.AddCommand(newConfigCmd())

	return ctlCmd
}

func getClient(cmd *cobra.Command) (*ravelctl.Client, error) {
	cfg, err := ravelctl.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	url := apiURL
	if url == "" {
		url = cfg.APIURL
	}

	tok := token
	if tok == "" {
		tok = cfg.Token
	}

	return ravelctl.NewClient(url, tok), nil
}
