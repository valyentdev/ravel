package server

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/internal/server"
	"github.com/valyentdev/ravel/pkg/config"
)

func init() {
	StartApiServerCmd.Flags().String("env", ".env", "Path to .env file")
	StartApiServerCmd.Flags().StringP("config", "c", "ravel.json", "Path to config file")
	ServerCmd.AddCommand(StartApiServerCmd)
}

var StartApiServerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API server",
	Long:  `Start the API server.`,
	Run: func(cmd *cobra.Command, args []string) {
		env := cmd.Flag("env").Value.String()
		err := godotenv.Load(env)
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		configFile := cmd.Flag("config").Value.String()

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			slog.Error("config file does not exist")
			os.Exit(1)
		}

		ravelConfig, err := config.ReadFile(configFile)
		if err != nil {
			slog.Error("failed to read config file", "error", err)
		}
		s, err := server.NewServer(ravelConfig)
		if err != nil {
			panic(err)
		}

		err = s.Serve()
		if err != nil {
			panic(err)
		}

	},
}
