package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/amirkh8006/bootup-cli/internal/iran"
	"github.com/spf13/cobra"
)

var iranCmd = &cobra.Command{
	Use:   "iran",
	Short: "Configure Iranian mirrors for package managers",
	Long:  `Interactive TUI to set Iranian mirrors for apt, npm, pip, Docker, and Go modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := iran.RunTUI(); err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}
	},
}

var iranListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available Iranian mirrors",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Iranian Mirrors")
		fmt.Println()
		for _, cat := range iran.Categories {
			fmt.Printf("📦 %s\n", cat.Name)
			fmt.Printf("   Current: %s\n", cat.Status())
			fmt.Println()
			for i, m := range cat.Mirrors {
				fmt.Printf("   %d. %-20s %s\n", i+1, m.Name, m.URL)
				fmt.Printf("      %s\n", m.Description)
			}
			fmt.Println()
		}
	},
}

var iranStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current active mirrors",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🇮🇷 Current Mirror Status")
		fmt.Println()
		for _, cat := range iran.Categories {
			fmt.Printf("  %-35s %s\n", cat.Name, cat.Status())
		}
		fmt.Println()
	},
}

var iranResetCmd = &cobra.Command{
	Use:   "reset [category]",
	Short: "Reset mirrors to defaults (all or one of: apt/npm/pip/docker/golang)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Print("Reset ALL mirrors to defaults? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(input)) != "y" {
				fmt.Println("Cancelled")
				return
			}
			for _, cat := range iran.Categories {
				fmt.Printf("Resetting %s...\n", cat.Name)
				if err := cat.Reset(); err != nil {
					fmt.Printf("❌ %s: %v\n", cat.Key, err)
				} else {
					fmt.Printf("✅ %s reset to default\n", cat.Key)
				}
			}
			return
		}

		key := args[0]
		for _, cat := range iran.Categories {
			if cat.Key == key {
				if err := cat.Reset(); err != nil {
					fmt.Printf("❌ Failed to reset %s: %v\n", cat.Key, err)
				} else {
					fmt.Printf("✅ %s reset to default\n", cat.Key)
				}
				return
			}
		}
		fmt.Printf("Unknown category: %s\nValid: apt, npm, pip, docker, golang\n", key)
	},
}

func init() {
	rootCmd.AddCommand(iranCmd)
	iranCmd.AddCommand(iranListCmd)
	iranCmd.AddCommand(iranStatusCmd)
	iranCmd.AddCommand(iranResetCmd)
}
