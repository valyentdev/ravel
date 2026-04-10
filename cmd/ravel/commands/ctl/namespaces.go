package ctl

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newNamespacesCmd() *cobra.Command {
	nsCmd := &cobra.Command{
		Use:     "namespaces",
		Aliases: []string{"ns", "namespace"},
		Short:   "Manage namespaces",
	}

	nsCmd.AddCommand(newNamespacesListCmd())
	nsCmd.AddCommand(newNamespacesCreateCmd())
	nsCmd.AddCommand(newNamespacesDeleteCmd())

	return nsCmd
}

func newNamespacesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List namespaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			namespaces, err := client.ListNamespaces()
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(namespaces, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCREATED")
			for _, ns := range namespaces {
				fmt.Fprintf(w, "%s\t%s\n", ns.Name, ns.CreatedAt)
			}
			w.Flush()

			return nil
		},
	}
}

func newNamespacesCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a namespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			ns, err := client.CreateNamespace(args[0])
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(ns, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Namespace %s created\n", ns.Name)
			return nil
		},
	}
}

func newNamespacesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a namespace",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			if err := client.DeleteNamespace(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Namespace %s deleted\n", args[0])
			return nil
		},
	}
}
