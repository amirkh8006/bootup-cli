package cmd

import (
	"fmt"
	"os"

	"github.com/amirkh8006/bootup-cli/internal/services"
	"github.com/amirkh8006/bootup-cli/internal/tui"
	"github.com/spf13/cobra"
)

// Version will be set during build time
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "bootup",
	Short:   "Bootup is a server setup CLI tool",
	Long:    `Bootup helps you install and configure common server apps and tools.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Run(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

var listServicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List available services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available services:")
		fmt.Println("To install a service, use the command: `bootup install [service]`")
		for _, service := range services.GetAllServices() {
			fmt.Printf(" - %s: %s\n", service.Name, service.Description)
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install [service]",
	Short: "Install a service",
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return services.GetServiceNames(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]

		installer, err := services.GetServiceInstaller(service)
		if err != nil {
			fmt.Printf("Service %s is not supported yet\n", service)
			os.Exit(1)
		}

		if err := installer(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(listServicesCmd)
	rootCmd.AddCommand(installCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
