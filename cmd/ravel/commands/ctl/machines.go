package ctl

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newMachinesCmd() *cobra.Command {
	machinesCmd := &cobra.Command{
		Use:     "machines",
		Aliases: []string{"m", "machine"},
		Short:   "Manage machines",
	}

	machinesCmd.AddCommand(newMachinesListCmd())
	machinesCmd.AddCommand(newMachinesGetCmd())
	machinesCmd.AddCommand(newMachinesLogsCmd())
	machinesCmd.AddCommand(newMachinesStartCmd())
	machinesCmd.AddCommand(newMachinesStopCmd())
	machinesCmd.AddCommand(newMachinesDeleteCmd())

	return machinesCmd
}

func newMachinesListCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List machines in a fleet",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			machines, err := client.ListMachines(namespace, fleet)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(machines, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tREGION\tSTATUS\tCREATED")
			for _, m := range machines {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Id, m.Region, m.Status, m.CreatedAt)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newMachinesGetCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:   "get <machine-id>",
		Short: "Get machine details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			machine, err := client.GetMachine(namespace, fleet, args[0])
			if err != nil {
				return err
			}

			data, _ := json.MarshalIndent(machine, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(data))

			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newMachinesLogsCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:   "logs <machine-id>",
		Short: "Get machine logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			logs, err := client.GetMachineLogs(namespace, fleet, args[0])
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), logs)

			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newMachinesStartCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:   "start <machine-id>",
		Short: "Start a machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			if err := client.StartMachine(namespace, fleet, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Machine %s started\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newMachinesStopCmd() *cobra.Command {
	var fleet string

	cmd := &cobra.Command{
		Use:   "stop <machine-id>",
		Short: "Stop a machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			if err := client.StopMachine(namespace, fleet, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Machine %s stopped\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.MarkFlagRequired("fleet")

	return cmd
}

func newMachinesDeleteCmd() *cobra.Command {
	var fleet string
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <machine-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a machine",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fleet == "" {
				return fmt.Errorf("--fleet is required")
			}

			client, err := getClient(cmd)
			if err != nil {
				return err
			}

			if err := client.DeleteMachine(namespace, fleet, args[0], force); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Machine %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVarP(&fleet, "fleet", "f", "", "Fleet name (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Force delete")
	cmd.MarkFlagRequired("fleet")

	return cmd
}
