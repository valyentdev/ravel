package ctl

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newFleetsCmd() *cobra.Command {
	fleetsCmd := &cobra.Command{
		Use:     "fleets",
		Aliases: []string{"f", "fleet"},
		Short:   "Manage fleets",
	}

	fleetsCmd.AddCommand(newFleetsListCmd())
	fleetsCmd.AddCommand(newFleetsCreateCmd())
	fleetsCmd.AddCommand(newFleetsDeleteCmd())

	return fleetsCmd
}

func newFleetsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List fleets",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			fleets, err := client.ListFleets(namespace)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(fleets, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tNAMESPACE\tCREATED")
			for _, f := range fleets {
				fmt.Fprintf(w, "%s\t%s\t%s\n", f.Name, f.Namespace, f.CreatedAt)
			}
			w.Flush()

			return nil
		},
	}
}

func newFleetsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a fleet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			fleet, err := client.CreateFleet(namespace, args[0])
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(fleet, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Fleet %s created\n", fleet.Name)
			return nil
		},
	}
}

func newFleetsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a fleet",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			if err := client.DeleteFleet(namespace, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Fleet %s deleted\n", args[0])
			return nil
		},
	}
}
