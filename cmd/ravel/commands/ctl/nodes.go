package ctl

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newNodesCmd() *cobra.Command {
	nodesCmd := &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"node"},
		Short:   "Manage nodes",
	}

	nodesCmd.AddCommand(newNodesListCmd())

	return nodesCmd
}

func newNodesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			nodes, err := client.ListNodes()
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(nodes, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tADDRESS\tREGION")
			for _, n := range nodes {
				fmt.Fprintf(w, "%s\t%s\t%s\n", n.Id, n.Address, n.Region)
			}
			w.Flush()

			return nil
		},
	}
}
