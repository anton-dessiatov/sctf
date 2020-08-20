package cmd

import (
	"fmt"
	"os"

	"github.com/anton-dessiatov/sctf/tf/cmd/cluster"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tf",
	Short: "Terraform usage for cloud provisioning demo",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(cluster.ClusterCmd)
}
