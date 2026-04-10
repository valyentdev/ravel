package ctl

import (
	"encoding/json"
	"fmt"

	"github.com/alexisbouchez/ravel/pkg/ravelctl"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ravelctl configuration",
	}

	configCmd.AddCommand(newConfigShowCmd())
	configCmd.AddCommand(newConfigSetCmd())

	return configCmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := ravelctl.LoadConfig()
			if err != nil {
				return err
			}

			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			fmt.Fprintf(cmd.OutOrStdout(), "\nConfig file: %s\n", ravelctl.DefaultConfigPath())
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	var url, tok string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := ravelctl.LoadConfig()
			if err != nil {
				return err
			}

			if url != "" {
				cfg.APIURL = url
			}
			if tok != "" {
				cfg.Token = tok
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Configuration saved to %s\n", ravelctl.DefaultConfigPath())
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "api", "", "API server URL")
	cmd.Flags().StringVar(&tok, "token", "", "API token")

	return cmd
}
