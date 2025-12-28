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

var servicesList = []string{"nginx"}


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
    Run: func(cmd *cobra.Command, args []string) {
        service := args[0]

        switch service {
        case "nginx":
            if err := services.InstallNginx(); err != nil {
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
