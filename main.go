package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mmahmoudian/tmp-place/cmd/admin"
	"github.com/mmahmoudian/tmp-place/cmd/janitor"
	"github.com/mmahmoudian/tmp-place/cmd/server"
	"github.com/mmahmoudian/tmp-place/cmd/setup"
)

func main() {
	//root of the CLI application
	var rootCmd = &cobra.Command{
		Use:   "tmp-place",
		Short: "tmp-place is a tool for admin tasks",
	}

	// server subcommand
	var ServerCmd = &cobra.Command{
		Use:   "server",
		Short: "Start a webserver that prints incoming requests",
		Run: func(cmd *cobra.Command, args []string) {
			server.ServerHandler(cmd, args)
		},
	}
	rootCmd.AddCommand(ServerCmd)

	// setup subcommand
	var SetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Set up the server",
		Run: func(cmd *cobra.Command, args []string) {
			setup.SetupHandler(cmd, args)
		},
	}
	rootCmd.AddCommand(SetupCmd)

	// admin subcommand
	var AdminCmd = &cobra.Command{
		Use:   "admin",
		Short: "Perform admin tasks",
		Run: func(cmd *cobra.Command, args []string) {
			admin.AdminHandler(cmd, args)
		},
	}
	rootCmd.AddCommand(AdminCmd)

	// janitor subcommand
	var janitorCmd = &cobra.Command{
		Use:   "janitor",
		Short: "Perform cleanup tasks for the uploaded files",
		Run: func(cmd *cobra.Command, args []string) {
			janitor.JanitorHandler(cmd, args)
		},
	}
	rootCmd.AddCommand(janitorCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
