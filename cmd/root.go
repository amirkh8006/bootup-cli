package cmd

import (
	"fmt"
	"os"

	"github.com/amirkh8006/bootup-cli/internal/services"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bootup",
	Short: "Bootup is a server setup CLI tool",
	Long:  `Bootup helps you install and configure common server apps and tools.`,
}

var servicesList = []string{"nginx", "postgresql", "mongodb", "redis", "nodejs", "kafka", "prometheus", "grafana", "alertmanager"}

var listServicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List available services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available services:")
		fmt.Println("To install a service, use the command: `bootup install [service]`")
		for _, service := range servicesList {
			fmt.Printf(" - %s\n", service)
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
		return servicesList, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]

		switch service {
		case "nginx":
			if err := services.InstallNginx(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "postgresql":
			if err := services.InstallPostgreSQL(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "mongodb":
			if err := services.InstallMongoDB(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "redis":
			if err := services.InstallRedis(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "nodejs":
			if err := services.InstallNodeJS(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "kafka":
			if err := services.InstallKafka(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "prometheus":
			if err := services.InstallPrometheus(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "grafana":
			if err := services.InstallGrafana(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "alertmanager":
			if err := services.InstallAlertmanager(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Service %s is not supported yet\n", service)
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
