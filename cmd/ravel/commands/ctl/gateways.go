package ctl

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newGatewaysCmd() *cobra.Command {
	gwCmd := &cobra.Command{
		Use:     "gateways",
		Aliases: []string{"gw", "gateway"},
		Short:   "Manage gateways",
	}

	gwCmd.AddCommand(newGatewaysListCmd())
	gwCmd.AddCommand(newGatewaysCreateCmd())
	gwCmd.AddCommand(newGatewaysDeleteCmd())

	return gwCmd
}

func newGatewaysListCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List gateways in a fleet",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			gateways, err := client.ListGateways(namespace, fleet)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(gateways, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTARGET PORT\tPROTOCOL")
			for _, gw := range gateways {
				fmt.Fprintf(w, "%s\t%d\t%s\n", gw.Name, gw.TargetPort, gw.Protocol)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newGatewaysCreateCmd() *cobra.Command {
	var fleet string
	var targetPort int

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			gw, err := client.CreateGateway(namespace, fleet, args[0], targetPort)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(gw, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Gateway %s created\n", gw.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.Flags().IntVarP(&targetPort, "port", "p", 80, "Target port")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newGatewaysDeleteCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a gateway",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			if err := client.DeleteGateway(namespace, fleet, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Gateway %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}
